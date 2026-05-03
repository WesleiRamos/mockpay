package tests

import (
	"testing"

	"github.com/wesleiramos/mockpay/internal/domain"
	"github.com/wesleiramos/mockpay/internal/service"
	"github.com/wesleiramos/mockpay/internal/store"
)

func newTestPayoutEnv() (*store.MemoryStore, *service.PixPayoutService) {
	s := store.NewWithPath(":memory:")
	ws := service.NewWebhookService(s, "", "")
	ps := service.NewPixPayoutService(s, ws)
	return s, ps
}

func TestPixPayoutCreate(t *testing.T) {
	_, ps := newTestPayoutEnv()

	payout, err := ps.CreatePayout(domain.CreatePixPayoutRequest{
		Amount:         5000,
		PixKeyType:     "cpf",
		PixKey:         "12345678901",
		ExternalID:     "ext-123",
		IdempotencyKey: "idem-key-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payout.Amount != 5000 {
		t.Errorf("expected 5000, got %d", payout.Amount)
	}
	if payout.Status != domain.PixPayoutProcessing {
		t.Errorf("expected PROCESSING, got %s", payout.Status)
	}
	if payout.PixKeyType != "cpf" {
		t.Errorf("expected cpf, got %s", payout.PixKeyType)
	}
}

func TestPixPayoutCreateIdempotent(t *testing.T) {
	_, ps := newTestPayoutEnv()

	req := domain.CreatePixPayoutRequest{
		Amount:         5000,
		PixKeyType:     "cpf",
		PixKey:         "12345678901",
		ExternalID:     "ext-123",
		IdempotencyKey: "idem-key-2",
	}

	p1, _ := ps.CreatePayout(req)
	p2, err := ps.CreatePayout(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p1.ID != p2.ID {
		t.Error("expected same payout id for idempotent request")
	}
	if p2.Status != domain.PixPayoutProcessing {
		t.Errorf("expected PROCESSING, got %s", p2.Status)
	}
}

func TestPixPayoutCreateInvalidKeyType(t *testing.T) {
	_, ps := newTestPayoutEnv()

	_, err := ps.CreatePayout(domain.CreatePixPayoutRequest{
		Amount:         5000,
		PixKeyType:     "invalid",
		PixKey:         "12345678901",
		IdempotencyKey: "idem-key-3",
	})
	if err == nil {
		t.Error("expected error for invalid pix_key_type")
	}
}

func TestPixPayoutCreateMissingKey(t *testing.T) {
	_, ps := newTestPayoutEnv()

	_, err := ps.CreatePayout(domain.CreatePixPayoutRequest{
		Amount:         5000,
		PixKeyType:     "cpf",
		PixKey:         "",
		IdempotencyKey: "idem-key-4",
	})
	if err == nil {
		t.Error("expected error for missing pix_key")
	}
}

func TestPixPayoutCreateMissingIdempotencyKey(t *testing.T) {
	_, ps := newTestPayoutEnv()

	_, err := ps.CreatePayout(domain.CreatePixPayoutRequest{
		Amount:     5000,
		PixKeyType: "cpf",
		PixKey:     "12345678901",
	})
	if err == nil {
		t.Error("expected error for missing idempotency_key")
	}
}

func TestPixPayoutCreateZeroAmount(t *testing.T) {
	_, ps := newTestPayoutEnv()

	_, err := ps.CreatePayout(domain.CreatePixPayoutRequest{
		Amount:         0,
		PixKeyType:     "cpf",
		PixKey:         "12345678901",
		IdempotencyKey: "idem-key-5",
	})
	if err == nil {
		t.Error("expected error for zero amount")
	}
}

func TestPixPayoutGet(t *testing.T) {
	_, ps := newTestPayoutEnv()

	payout, _ := ps.CreatePayout(domain.CreatePixPayoutRequest{
		Amount:         3000,
		PixKeyType:     "email",
		PixKey:         "test@example.com",
		IdempotencyKey: "idem-key-6",
	})

	got, err := ps.GetPayout(payout.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Amount != 3000 {
		t.Errorf("expected 3000, got %d", got.Amount)
	}
}

func TestPixPayoutGetNotFound(t *testing.T) {
	_, ps := newTestPayoutEnv()

	_, err := ps.GetPayout("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent payout")
	}
}

func TestPixPayoutValidKeyTypes(t *testing.T) {
	_, ps := newTestPayoutEnv()

	keyTypes := []string{"cpf", "phone", "email", "random"}
	for _, kt := range keyTypes {
		_, err := ps.CreatePayout(domain.CreatePixPayoutRequest{
			Amount:         1000,
			PixKeyType:     kt,
			PixKey:         "test-key",
			IdempotencyKey: "ik-" + kt,
		})
		if err != nil {
			t.Errorf("expected no error for key type %s, got: %v", kt, err)
		}
	}
}

func TestPixPayoutStoreCRUD(t *testing.T) {
	s := store.NewWithPath(":memory:")

	p := &domain.PixPayout{
		ID:             "payout_1",
		Amount:         5000,
		Status:         domain.PixPayoutProcessing,
		ExternalID:     "ext-1",
		EndToEndID:     "",
		PixKeyType:     "cpf",
		PixKey:         "12345678901",
		IdempotencyKey: "ik-1",
		Metadata:       map[string]string{"source": "wallet"},
	}
	s.CreatePixPayout(p)

	got, ok := s.GetPixPayout("payout_1")
	if !ok {
		t.Fatal("expected payout to be found")
	}
	if got.Amount != 5000 {
		t.Errorf("expected 5000, got %d", got.Amount)
	}
	if got.Metadata["source"] != "wallet" {
		t.Errorf("expected source=wallet, got %s", got.Metadata["source"])
	}

	byKey, ok := s.GetPixPayoutByIdempotencyKey("ik-1")
	if !ok {
		t.Fatal("expected payout to be found by idempotency key")
	}
	if byKey.ID != "payout_1" {
		t.Errorf("expected payout_1, got %s", byKey.ID)
	}

	s.UpdatePixPayoutStatus("payout_1", domain.PixPayoutLiquidated, "E2026050212000000000000000000001")
	got, _ = s.GetPixPayout("payout_1")
	if got.Status != domain.PixPayoutLiquidated {
		t.Errorf("expected LIQUIDATED, got %s", got.Status)
	}
	if got.EndToEndID != "E2026050212000000000000000000001" {
		t.Errorf("expected end_to_end_id, got '%s'", got.EndToEndID)
	}
}
