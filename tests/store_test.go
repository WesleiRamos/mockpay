package tests

import (
	"fmt"
	"sync"
	"testing"

	"github.com/wesleiramos/mockpay/internal/domain"
	"github.com/wesleiramos/mockpay/internal/store"
)

func newTestStore() *store.MemoryStore {
	return store.NewWithPath(":memory:")
}

func TestCreateAndGetBilling(t *testing.T) {
	s := newTestStore()
	b := &domain.Billing{
		ID:     "bill_test123",
		Amount: 5000,
		Status: domain.StatusPending,
	}

	s.CreateBilling(b)

	got, ok := s.GetBilling("bill_test123")
	if !ok {
		t.Fatal("expected billing to be found")
	}
	if got.Amount != 5000 {
		t.Errorf("expected amount 5000, got %d", got.Amount)
	}
}

func TestGetBillingNotFound(t *testing.T) {
	s := newTestStore()
	_, ok := s.GetBilling("nonexistent")
	if ok {
		t.Error("expected billing not to be found")
	}
}

func TestListBillings(t *testing.T) {
	s := newTestStore()
	s.CreateBilling(&domain.Billing{ID: "bill_1", Amount: 1000})
	s.CreateBilling(&domain.Billing{ID: "bill_2", Amount: 2000})

	list := s.ListBillings()
	if len(list) != 2 {
		t.Errorf("expected 2 billings, got %d", len(list))
	}
}

func TestUpdateBillingStatus(t *testing.T) {
	s := newTestStore()
	s.CreateBilling(&domain.Billing{ID: "bill_1", Status: domain.StatusPending})

	s.UpdateBillingStatus("bill_1", domain.StatusApproved)

	got, _ := s.GetBilling("bill_1")
	if got.Status != domain.StatusApproved {
		t.Errorf("expected APPROVED, got %s", got.Status)
	}
}

func TestCustomerCRUD(t *testing.T) {
	s := newTestStore()

	c := &domain.Customer{
		ID:       "cust_1",
		Metadata: map[string]string{"email": "test@test.com", "name": "Test"},
	}
	s.CreateCustomer(c)

	got, ok := s.GetCustomer("cust_1")
	if !ok {
		t.Fatal("expected customer to be found")
	}
	if got.Metadata["email"] != "test@test.com" {
		t.Errorf("expected email test@test.com, got %s", got.Metadata["email"])
	}

	byEmail, ok := s.GetCustomerByEmail("test@test.com")
	if !ok {
		t.Fatal("expected customer to be found by email")
	}
	if byEmail.ID != "cust_1" {
		t.Errorf("expected cust_1, got %s", byEmail.ID)
	}

	list := s.ListCustomers()
	if len(list) != 1 {
		t.Errorf("expected 1 customer, got %d", len(list))
	}
}

func TestCouponCRUD(t *testing.T) {
	s := newTestStore()

	c := &domain.Coupon{
		ID:           "cpn_1",
		Code:         "TEST10",
		DiscountKind: domain.DiscountPercentage,
		Discount:     10,
		MaxRedeems:   5,
		Status:       domain.CouponActive,
	}
	s.CreateCoupon(c)

	got, ok := s.GetCoupon("cpn_1")
	if !ok {
		t.Fatal("expected coupon to be found")
	}
	if got.Code != "TEST10" {
		t.Errorf("expected TEST10, got %s", got.Code)
	}

	s.IncrementCouponRedeems("cpn_1")
	got, _ = s.GetCoupon("cpn_1")
	if got.RedeemsCount != 1 {
		t.Errorf("expected 1 redeem, got %d", got.RedeemsCount)
	}
}

func TestPixCRUD(t *testing.T) {
	s := newTestStore()

	p := &domain.PixCharge{
		ID:     "pix_1",
		Amount: 3000,
		Status: domain.PixPending,
	}
	s.CreatePixCharge(p)

	got, ok := s.GetPixCharge("pix_1")
	if !ok {
		t.Fatal("expected pix charge to be found")
	}
	if got.Amount != 3000 {
		t.Errorf("expected 3000, got %d", got.Amount)
	}

	s.UpdatePixStatus("pix_1", domain.PixApproved)
	got, _ = s.GetPixCharge("pix_1")
	if got.Status != domain.PixApproved {
		t.Errorf("expected APPROVED, got %s", got.Status)
	}
}

func TestStats(t *testing.T) {
	s := newTestStore()
	s.CreateBilling(&domain.Billing{ID: "b1", Status: domain.StatusPending})
	s.CreateBilling(&domain.Billing{ID: "b2", Status: domain.StatusApproved})
	s.CreatePixCharge(&domain.PixCharge{ID: "p1", Status: domain.PixPending})

	stats := s.Stats()
	if stats["billings_total"] != 2 {
		t.Errorf("expected 2 billings, got %d", stats["billings_total"])
	}
	if stats["billings_pending"] != 1 {
		t.Errorf("expected 1 pending, got %d", stats["billings_pending"])
	}
	if stats["billings_approved"] != 1 {
		t.Errorf("expected 1 approved, got %d", stats["billings_approved"])
	}
	if stats["pix_total"] != 1 {
		t.Errorf("expected 1 pix, got %d", stats["pix_total"])
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := newTestStore()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			s.CreateBilling(&domain.Billing{
				ID:     fmt.Sprintf("bill_%d", n),
				Amount: int64(n * 100),
			})
		}(i)
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.ListBillings()
		}()
	}

	wg.Wait()

	list := s.ListBillings()
	if len(list) != 100 {
		t.Errorf("expected 100 billings, got %d", len(list))
	}
}

func TestClearNextBilling(t *testing.T) {
	s := newTestStore()
	nb := "2026-06-01T12:00:00.000"
	s.CreateBilling(&domain.Billing{
		ID:          "bill_recur_1",
		Status:      domain.StatusPending,
		Frequency:   domain.FrequencyMultipleBilling,
		NextBilling: &nb,
	})

	got, _ := s.GetBilling("bill_recur_1")
	if got.NextBilling == nil {
		t.Fatal("expected next_billing to be set")
	}

	s.ClearNextBilling("bill_recur_1")

	got, _ = s.GetBilling("bill_recur_1")
	if got.NextBilling != nil {
		t.Error("expected next_billing to be nil after clear")
	}
}

func TestFindByID(t *testing.T) {
	s := newTestStore()
	s.CreateBilling(&domain.Billing{ID: "bill_1"})
	s.CreatePixCharge(&domain.PixCharge{ID: "pix_1"})

	result, ok := s.FindByID("bill_1")
	if !ok {
		t.Fatal("expected to find billing")
	}
	if _, isBilling := result.(*domain.Billing); !isBilling {
		t.Error("expected billing type")
	}

	result, ok = s.FindByID("pix_1")
	if !ok {
		t.Fatal("expected to find pix charge")
	}
	if _, isPix := result.(*domain.PixCharge); !isPix {
		t.Error("expected pix charge type")
	}

	_, ok = s.FindByID("nonexistent")
	if ok {
		t.Error("expected not to find anything")
	}
}
