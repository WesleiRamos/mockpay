package service

import (
	"fmt"
	"log"
	"time"

	"github.com/wesleiramos/mockpay/internal/domain"
	"github.com/wesleiramos/mockpay/internal/store"
	"github.com/wesleiramos/mockpay/internal/util"
)

type BillingService struct {
	store               *store.MemoryStore
	baseURL             string
	webhook             *WebhookService
	couponSvc           *CouponService
	defaultInterestRate float64
}

func NewBillingService(s *store.MemoryStore, baseURL string, w *WebhookService, cs *CouponService, defaultInterestRate float64) *BillingService {
	return &BillingService{store: s, baseURL: baseURL, webhook: w, couponSvc: cs, defaultInterestRate: defaultInterestRate}
}

func (bs *BillingService) Create(req domain.CreateBillingRequest) (*domain.Billing, error) {
	if len(req.Products) == 0 {
		return nil, fmt.Errorf("at least one product is required")
	}

	var total int64
	for _, p := range req.Products {
		if p.Quantity < 1 {
			return nil, fmt.Errorf("product quantity must be >= 1")
		}
		if p.Price < 100 {
			return nil, fmt.Errorf("product price must be >= 100 cents")
		}
		total += p.Price * int64(p.Quantity)
	}

	installments := req.Installments
	if installments < 1 {
		installments = 1
	}

	interestRate := req.InterestRate
	if interestRate == 0 {
		interestRate = bs.defaultInterestRate
	}

	id := util.NewID()
	now := domain.NowTimestamp()

	var customerRef *domain.CustomerRef
	if req.CustomerID != "" {
		if c, ok := bs.store.GetCustomer(req.CustomerID); ok {
			customerRef = &domain.CustomerRef{ID: c.ID, Metadata: c.Metadata}
		}
	} else if req.Customer != nil {
		c := &domain.Customer{
			ID:        util.NewID(),
			Metadata:  map[string]string{
				"name":      req.Customer.Name,
				"email":     req.Customer.Email,
				"cellphone": req.Customer.Cellphone,
				"tax_id":    req.Customer.TaxID,
			},
			CreatedAt: now,
			UpdatedAt: now,
		}
		bs.store.CreateCustomer(c)
		customerRef = &domain.CustomerRef{ID: c.ID, Metadata: c.Metadata}
	}

	var nextBilling *string
	if req.Frequency == domain.FrequencyMultipleBilling {
		nb := domain.FutureTimestamp(30)
		nextBilling = &nb
	}

	finalAmount := total
	var originalAmount int64
	var couponCode string

	if req.CouponCode != "" {
		discounted, err := bs.couponSvc.ApplyDiscountByCode(req.CouponCode, total)
		if err != nil {
			return nil, fmt.Errorf("coupon error: %v", err)
		}
		originalAmount = total
		finalAmount = discounted
		couponCode = req.CouponCode
	}

	billing := &domain.Billing{
		ID:            id,
		URL:           fmt.Sprintf("%s/checkout/%s", bs.baseURL, id),
		Amount:        finalAmount,
		OriginalAmount: originalAmount,
		CouponCode:    couponCode,
		Status:        domain.StatusPending,
		DevMode:       true,
		Methods:       req.Methods,
		Products:      req.Products,
		Frequency:     req.Frequency,
		NextBilling:   nextBilling,
		Customer:      customerRef,
		ReturnURL:     req.ReturnURL,
		CompletionURL: req.CompletionURL,
		Installments:  installments,
		InterestRate:  interestRate,
		InstallmentList: domain.CalculateInstallments(finalAmount, installments, interestRate),
		ExternalID:    req.ExternalID,
		Metadata:      req.Metadata,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	bs.store.CreateBilling(billing)
	return billing, nil
}

func (bs *BillingService) Get(id string) (*domain.Billing, error) {
	b, ok := bs.store.GetBilling(id)
	if !ok {
		return nil, fmt.Errorf("billing not found")
	}
	return b, nil
}

func (bs *BillingService) List() []*domain.Billing {
	return bs.store.ListBillings()
}

func (bs *BillingService) Approve(id string) (*domain.Billing, error) {
	b, ok := bs.store.GetBilling(id)
	if !ok {
		return nil, fmt.Errorf("billing not found")
	}
	if b.Status != domain.StatusPending {
		return nil, fmt.Errorf("billing is not pending")
	}

	bs.store.UpdateBillingStatus(id, domain.StatusApproved)

	if len(b.InstallmentList) > 0 {
		now := domain.NowTimestamp()
		for i := range b.InstallmentList {
			b.InstallmentList[i].Status = string(domain.StatusApproved)
			b.InstallmentList[i].PaidAt = &now
		}
		bs.store.UpdateBillingInstallments(id, b.InstallmentList)
	}

	updated, _ := bs.store.GetBilling(id)
	bs.webhook.Dispatch(domain.EventBillingApproved, updated)

	return updated, nil
}

func (bs *BillingService) Deny(id string) (*domain.Billing, error) {
	b, ok := bs.store.GetBilling(id)
	if !ok {
		return nil, fmt.Errorf("billing not found")
	}
	if b.Status != domain.StatusPending {
		return nil, fmt.Errorf("billing is not pending")
	}

	bs.store.UpdateBillingStatus(id, domain.StatusDenied)

	if len(b.InstallmentList) > 0 {
		for i := range b.InstallmentList {
			b.InstallmentList[i].Status = string(domain.StatusDenied)
		}
		bs.store.UpdateBillingInstallments(id, b.InstallmentList)
	}

	updated, _ := bs.store.GetBilling(id)
	bs.webhook.Dispatch(domain.EventBillingDenied, updated)

	return updated, nil
}

func (bs *BillingService) Cancel(id string) (*domain.Billing, error) {
	b, ok := bs.store.GetBilling(id)
	if !ok {
		return nil, fmt.Errorf("billing not found")
	}

	if b.Status == domain.StatusCancelled {
		return nil, fmt.Errorf("billing is already cancelled")
	}

	if b.Frequency == domain.FrequencyMultipleBilling {
		if b.Status != domain.StatusPending && b.Status != domain.StatusApproved {
			return nil, fmt.Errorf("only pending or approved recurring billings can be cancelled")
		}
		bs.store.UpdateBillingStatus(id, domain.StatusCancelled)
		bs.store.ClearNextBilling(id)
	} else {
		if b.Status != domain.StatusPending {
			return nil, fmt.Errorf("only pending billings can be cancelled")
		}
		bs.store.UpdateBillingStatus(id, domain.StatusCancelled)
	}

	if len(b.InstallmentList) > 0 {
		for i := range b.InstallmentList {
			b.InstallmentList[i].Status = string(domain.StatusCancelled)
		}
		bs.store.UpdateBillingInstallments(id, b.InstallmentList)
	}

	updated, _ := bs.store.GetBilling(id)
	bs.webhook.Dispatch(domain.EventBillingCancelled, updated)

	return updated, nil
}

func (bs *BillingService) GetInstallments(id string) ([]domain.Installment, error) {
	b, ok := bs.store.GetBilling(id)
	if !ok {
		return nil, fmt.Errorf("billing not found")
	}
	return b.InstallmentList, nil
}

func (bs *BillingService) CheckRecurring() {
	pending := bs.store.ListBillingsByStatus(domain.StatusPending)
	for _, b := range pending {
		if b.Frequency != domain.FrequencyMultipleBilling || b.NextBilling == nil {
			continue
		}

		nextTime, err := time.Parse("2006-01-02T15:04:05.000", *b.NextBilling)
		if err != nil {
			continue
		}

		if time.Now().UTC().After(nextTime) {
			newBilling, err := bs.Create(domain.CreateBillingRequest{
				Frequency:     domain.FrequencyMultipleBilling,
				Methods:       b.Methods,
				Products:      b.Products,
				ReturnURL:     b.ReturnURL,
				CompletionURL: b.CompletionURL,
				Installments:  b.Installments,
				InterestRate:  b.InterestRate,
			})
			if err != nil {
				log.Printf("error creating recurring billing: %v", err)
				continue
			}

			if b.Customer != nil {
				newBilling.Customer = b.Customer
			}

			nb := domain.FutureTimestamp(30)
			newBilling.NextBilling = &nb

			log.Printf("created recurring billing %s from parent %s", newBilling.ID, b.ID)
			bs.webhook.Dispatch(domain.EventBillingCreated, newBilling)
		}
	}
}
