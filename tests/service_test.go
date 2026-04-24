package tests

import (
	"testing"

	"github.com/wesleiramos/mockpay/internal/domain"
	"github.com/wesleiramos/mockpay/internal/service"
	"github.com/wesleiramos/mockpay/internal/store"
)

func newTestEnv() (*store.MemoryStore, *service.BillingService, *service.CustomerService, *service.CouponService, *service.PixService) {
	s := store.NewWithPath(":memory:")
	ws := service.NewWebhookService(s, "", "")
	cs := service.NewCustomerService(s)
	cps := service.NewCouponService(s)
	bs := service.NewBillingService(s, "http://localhost:8080", ws, cps, 0)
	ps := service.NewPixService(s, ws, "http://localhost:8080")
	return s, bs, cs, cps, ps
}

func TestBillingServiceCreate(t *testing.T) {
	_, bs, _, _, _ := newTestEnv()

	billing, err := bs.Create(domain.CreateBillingRequest{
		Frequency: domain.FrequencyOneTime,
		Methods:   []domain.BillingMethod{domain.MethodPIX},
		Products:  []domain.Product{
			{ExternalID: "p1", Name: "Test", Quantity: 1, Price: 5000},
		},
		ReturnURL:     "http://localhost:3000",
		CompletionURL: "http://localhost:3000/done",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if billing.Amount != 5000 {
		t.Errorf("expected 5000, got %d", billing.Amount)
	}
	if billing.Status != domain.StatusPending {
		t.Errorf("expected PENDING, got %s", billing.Status)
	}
	if billing.URL == "" {
		t.Error("expected checkout URL to be set")
	}
}

func TestBillingServiceCreateNoProducts(t *testing.T) {
	_, bs, _, _, _ := newTestEnv()

	_, err := bs.Create(domain.CreateBillingRequest{
		Frequency: domain.FrequencyOneTime,
		Methods:   []domain.BillingMethod{domain.MethodPIX},
		Products:  []domain.Product{},
	})
	if err == nil {
		t.Error("expected error for empty products")
	}
}

func TestBillingServiceCreateInvalidPrice(t *testing.T) {
	_, bs, _, _, _ := newTestEnv()

	_, err := bs.Create(domain.CreateBillingRequest{
		Products: []domain.Product{
			{ExternalID: "p1", Name: "Test", Quantity: 1, Price: 50},
		},
	})
	if err == nil {
		t.Error("expected error for price < 100")
	}
}

func TestBillingServiceApproveDeny(t *testing.T) {
	_, bs, _, _, _ := newTestEnv()

	billing, _ := bs.Create(domain.CreateBillingRequest{
		Products: []domain.Product{
			{ExternalID: "p1", Name: "Test", Quantity: 1, Price: 5000},
		},
	})

	approved, err := bs.Approve(billing.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if approved.Status != domain.StatusApproved {
		t.Errorf("expected APPROVED, got %s", approved.Status)
	}

	// Can't approve again
	_, err = bs.Approve(billing.ID)
	if err == nil {
		t.Error("expected error for re-approval")
	}
}

func TestBillingServiceWithInstallments(t *testing.T) {
	_, bs, _, _, _ := newTestEnv()

	billing, err := bs.Create(domain.CreateBillingRequest{
		Products: []domain.Product{
			{ExternalID: "p1", Name: "Test", Quantity: 1, Price: 10000},
		},
		Installments: 3,
		InterestRate: 5.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(billing.InstallmentList) != 3 {
		t.Fatalf("expected 3 installments, got %d", len(billing.InstallmentList))
	}

	total := int64(0)
	for _, inst := range billing.InstallmentList {
		total += inst.Amount
	}
	// 10000 * (1 + 5/100 * 3) = 11500
	if total != 11500 {
		t.Errorf("expected total 11500, got %d", total)
	}

	// Approve and check installments are approved
	bs.Approve(billing.ID)
	insts, _ := bs.GetInstallments(billing.ID)
	for _, inst := range insts {
		if inst.Status != string(domain.StatusApproved) {
			t.Errorf("expected APPROVED, got %s", inst.Status)
		}
	}
}

func TestBillingServiceRecurring(t *testing.T) {
	_, bs, _, _, _ := newTestEnv()

	billing, err := bs.Create(domain.CreateBillingRequest{
		Frequency: domain.FrequencyMultipleBilling,
		Products: []domain.Product{
			{ExternalID: "p1", Name: "Test", Quantity: 1, Price: 5000},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if billing.NextBilling == nil {
		t.Error("expected next_billing to be set for recurring")
	}
	if billing.Frequency != domain.FrequencyMultipleBilling {
		t.Errorf("expected MULTIPLE_PAYMENTS, got %s", billing.Frequency)
	}
}

func TestCustomerServiceCreate(t *testing.T) {
	_, _, cs, _, _ := newTestEnv()

	customer, err := cs.Create(domain.CustomerInput{
		Name:  "João",
		Email: "joao@test.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if customer.Metadata["email"] != "joao@test.com" {
		t.Errorf("expected joao@test.com, got %s", customer.Metadata["email"])
	}
}

func TestCustomerServiceCreateDuplicate(t *testing.T) {
	_, _, cs, _, _ := newTestEnv()

	cs.Create(domain.CustomerInput{Email: "joao@test.com"})
	c2, err := cs.Create(domain.CustomerInput{Email: "joao@test.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c2.Metadata["email"] != "joao@test.com" {
		t.Error("expected same customer returned")
	}
}

func TestCustomerServiceNoEmail(t *testing.T) {
	_, _, cs, _, _ := newTestEnv()

	_, err := cs.Create(domain.CustomerInput{Name: "João"})
	if err == nil {
		t.Error("expected error for missing email")
	}
}

func TestCouponServiceCreate(t *testing.T) {
	_, _, _, cps, _ := newTestEnv()

	coupon, err := cps.Create(domain.CreateCouponRequest{
		Code:         "TEST10",
		DiscountKind: domain.DiscountPercentage,
		Discount:     10,
		MaxRedeems:   5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if coupon.Code != "TEST10" {
		t.Errorf("expected TEST10, got %s", coupon.Code)
	}
}

func TestCouponServiceApplyDiscountPercentage(t *testing.T) {
	_, _, _, cps, _ := newTestEnv()

	coupon, _ := cps.Create(domain.CreateCouponRequest{
		Code:         "DESC20",
		DiscountKind: domain.DiscountPercentage,
		Discount:     20,
		MaxRedeems:   10,
	})

	result, err := cps.ApplyDiscount(coupon.ID, 10000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 8000 {
		t.Errorf("expected 8000, got %d", result)
	}
}

func TestCouponServiceApplyDiscountFixed(t *testing.T) {
	_, _, _, cps, _ := newTestEnv()

	coupon, _ := cps.Create(domain.CreateCouponRequest{
		Code:         "FLAT500",
		DiscountKind: domain.DiscountFixed,
		Discount:     500,
		MaxRedeems:   -1,
	})

	result, err := cps.ApplyDiscount(coupon.ID, 5000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 4500 {
		t.Errorf("expected 4500, got %d", result)
	}
}

func TestPixServiceCreate(t *testing.T) {
	_, _, _, _, ps := newTestEnv()

	charge, err := ps.Create(domain.CreatePixRequest{
		Amount:    5000,
		ExpiresIn: 3600,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if charge.Amount != 5000 {
		t.Errorf("expected 5000, got %d", charge.Amount)
	}
	if charge.Status != domain.PixPending {
		t.Errorf("expected PENDING, got %s", charge.Status)
	}
	if charge.BrCode == "" {
		t.Error("expected br_code to be set")
	}
}

func TestPixServiceCreateLowAmount(t *testing.T) {
	_, _, _, _, ps := newTestEnv()

	_, err := ps.Create(domain.CreatePixRequest{Amount: 50})
	if err == nil {
		t.Error("expected error for amount < 100")
	}
}

func TestBillingServiceCancelOneTime(t *testing.T) {
	_, bs, _, _, _ := newTestEnv()

	billing, _ := bs.Create(domain.CreateBillingRequest{
		Products: []domain.Product{
			{ExternalID: "p1", Name: "Test", Quantity: 1, Price: 5000},
		},
	})

	cancelled, err := bs.Cancel(billing.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cancelled.Status != domain.StatusCancelled {
		t.Errorf("expected CANCELLED, got %s", cancelled.Status)
	}
}

func TestBillingServiceCancelRecurring(t *testing.T) {
	_, bs, _, _, _ := newTestEnv()

	billing, _ := bs.Create(domain.CreateBillingRequest{
		Frequency: domain.FrequencyMultipleBilling,
		Products: []domain.Product{
			{ExternalID: "p1", Name: "Test", Quantity: 1, Price: 9900},
		},
	})

	if billing.NextBilling == nil {
		t.Fatal("expected next_billing to be set")
	}

	cancelled, err := bs.Cancel(billing.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cancelled.Status != domain.StatusCancelled {
		t.Errorf("expected CANCELLED, got %s", cancelled.Status)
	}
	if cancelled.NextBilling != nil {
		t.Error("expected next_billing to be nil after cancel")
	}
}

func TestBillingServiceCancelNotFound(t *testing.T) {
	_, bs, _, _, _ := newTestEnv()

	_, err := bs.Cancel("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent billing")
	}
}

func TestBillingServiceCancelAlreadyApproved(t *testing.T) {
	_, bs, _, _, _ := newTestEnv()

	billing, _ := bs.Create(domain.CreateBillingRequest{
		Products: []domain.Product{
			{ExternalID: "p1", Name: "Test", Quantity: 1, Price: 5000},
		},
	})
	bs.Approve(billing.ID)

	_, err := bs.Cancel(billing.ID)
	if err == nil {
		t.Error("expected error for cancelling approved one-time billing")
	}
}

func TestBillingServiceCancelRecurringAfterApproval(t *testing.T) {
	_, bs, _, _, _ := newTestEnv()

	billing, _ := bs.Create(domain.CreateBillingRequest{
		Frequency: domain.FrequencyMultipleBilling,
		Products: []domain.Product{
			{ExternalID: "p1", Name: "Test", Quantity: 1, Price: 9900},
		},
	})
	bs.Approve(billing.ID)

	cancelled, err := bs.Cancel(billing.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cancelled.Status != domain.StatusCancelled {
		t.Errorf("expected CANCELLED, got %s", cancelled.Status)
	}
	if cancelled.NextBilling != nil {
		t.Error("expected next_billing to be nil after cancel")
	}
}

func TestBillingServiceCancelAlreadyCancelled(t *testing.T) {
	_, bs, _, _, _ := newTestEnv()

	billing, _ := bs.Create(domain.CreateBillingRequest{
		Products: []domain.Product{
			{ExternalID: "p1", Name: "Test", Quantity: 1, Price: 5000},
		},
	})
	bs.Cancel(billing.ID)

	_, err := bs.Cancel(billing.ID)
	if err == nil {
		t.Error("expected error for cancelling already cancelled billing")
	}
}

func TestBillingServiceCancelWithInstallments(t *testing.T) {
	_, bs, _, _, _ := newTestEnv()

	billing, _ := bs.Create(domain.CreateBillingRequest{
		Products: []domain.Product{
			{ExternalID: "p1", Name: "Test", Quantity: 1, Price: 10000},
		},
		Installments: 3,
	})

	cancelled, err := bs.Cancel(billing.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cancelled.Status != domain.StatusCancelled {
		t.Errorf("expected CANCELLED, got %s", cancelled.Status)
	}

	insts, _ := bs.GetInstallments(billing.ID)
	for _, inst := range insts {
		if inst.Status != string(domain.StatusCancelled) {
			t.Errorf("expected CANCELLED installment, got %s", inst.Status)
		}
	}
}

func TestPixServiceApprove(t *testing.T) {
	_, _, _, _, ps := newTestEnv()

	charge, _ := ps.Create(domain.CreatePixRequest{Amount: 5000})
	approved, err := ps.Approve(charge.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if approved.Status != domain.PixApproved {
		t.Errorf("expected APPROVED, got %s", approved.Status)
	}
}
