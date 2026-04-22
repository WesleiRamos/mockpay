package service

import (
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/skip2/go-qrcode"
	"github.com/wesleiramos/mockpay/internal/domain"
	"github.com/wesleiramos/mockpay/internal/store"
	"github.com/wesleiramos/mockpay/internal/util"
)

type PixService struct {
	store   *store.MemoryStore
	webhook *WebhookService
	baseURL string
}

func NewPixService(s *store.MemoryStore, w *WebhookService, baseURL string) *PixService {
	return &PixService{store: s, webhook: w, baseURL: baseURL}
}

func (ps *PixService) Create(req domain.CreatePixRequest) (*domain.PixCharge, error) {
	if req.Amount < 100 {
		return nil, fmt.Errorf("amount must be >= 100 cents")
	}

	expiresIn := req.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}

	now := domain.NowTimestamp()
	expiresAt := time.Now().UTC().Add(time.Duration(expiresIn) * time.Second).Format("2006-01-02T15:04:05.000")

	var customerRef *domain.CustomerRef
	if req.Customer != nil {
		c := &domain.Customer{
			ID:       util.NewID(),
			Metadata: map[string]string{
				"name":      req.Customer.Name,
				"email":     req.Customer.Email,
				"cellphone": req.Customer.Cellphone,
				"tax_id":    req.Customer.TaxID,
			},
			CreatedAt: now,
			UpdatedAt: now,
		}
		ps.store.CreateCustomer(c)
		customerRef = &domain.CustomerRef{ID: c.ID, Metadata: c.Metadata}
	}

	id := util.NewID()

	qrURL := fmt.Sprintf("%s/checkout/%s/approve", ps.baseURL, id)
	png, err := qrcode.Encode(qrURL, qrcode.Medium, 256)
	if err != nil {
		log.Printf("qr code generation error: %v", err)
		png = nil
	}

	var brCodeBase64 string
	if png != nil {
		brCodeBase64 = "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	}

	charge := &domain.PixCharge{
		ID:           id,
		Amount:       req.Amount,
		Status:       domain.PixPending,
		DevMode:      true,
		BrCode:       qrURL,
		BrCodeBase64: brCodeBase64,
		PlatformFee:  80,
		ExpiresAt:    expiresAt,
		Customer:     customerRef,
		ExternalID:   req.ExternalID,
		Metadata:     req.Metadata,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	ps.store.CreatePixCharge(charge)
	return charge, nil
}

func (ps *PixService) Get(id string) (*domain.PixCharge, error) {
	p, ok := ps.store.GetPixCharge(id)
	if !ok {
		return nil, fmt.Errorf("pix charge not found")
	}
	return p, nil
}

func (ps *PixService) Approve(id string) (*domain.PixCharge, error) {
	p, ok := ps.store.GetPixCharge(id)
	if !ok {
		return nil, fmt.Errorf("pix charge not found")
	}
	if p.Status != domain.PixPending {
		return nil, fmt.Errorf("pix charge is not pending")
	}

	ps.store.UpdatePixStatus(id, domain.PixApproved)
	updated, _ := ps.store.GetPixCharge(id)
	ps.webhook.Dispatch(domain.EventPixApproved, updated)

	return updated, nil
}

func (ps *PixService) Deny(id string) (*domain.PixCharge, error) {
	p, ok := ps.store.GetPixCharge(id)
	if !ok {
		return nil, fmt.Errorf("pix charge not found")
	}
	if p.Status != domain.PixPending {
		return nil, fmt.Errorf("pix charge is not pending")
	}

	ps.store.UpdatePixStatus(id, domain.PixExpired)
	updated, _ := ps.store.GetPixCharge(id)

	return updated, nil
}

func (ps *PixService) CheckExpired() {
	charges := ps.store.ListPixCharges()
	for _, p := range charges {
		if p.Status != domain.PixPending {
			continue
		}

		expiresAt, err := time.Parse("2006-01-02T15:04:05.000", p.ExpiresAt)
		if err != nil {
			continue
		}

		if time.Now().UTC().After(expiresAt) {
			ps.store.UpdatePixStatus(p.ID, domain.PixExpired)
			ps.webhook.Dispatch(domain.EventPixExpired, p)
			log.Printf("pix charge %s expired", p.ID)
		}
	}
}
