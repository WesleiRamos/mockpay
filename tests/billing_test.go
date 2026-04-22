package tests

import (
	"testing"

	"github.com/wesleiramos/mockpay/internal/domain"
)

func TestCalculateInstallmentsSingle(t *testing.T) {
	result := domain.CalculateInstallments(10000, 1, 0)
	if result != nil {
		t.Error("expected nil for single installment")
	}
}

func TestCalculateInstallmentsNoInterest(t *testing.T) {
	result := domain.CalculateInstallments(10000, 2, 0)
	if len(result) != 2 {
		t.Fatalf("expected 2 installments, got %d", len(result))
	}

	total := result[0].Amount + result[1].Amount
	if total != 10000 {
		t.Errorf("expected total 10000, got %d", total)
	}

	if result[0].Number != 1 || result[1].Number != 2 {
		t.Error("expected installment numbers 1 and 2")
	}
}

func TestCalculateInstallmentsWithInterest(t *testing.T) {
	result := domain.CalculateInstallments(10000, 3, 2.5)
	if len(result) != 3 {
		t.Fatalf("expected 3 installments, got %d", len(result))
	}

	// Total = 10000 * (1 + 2.5/100 * 3) = 10000 * 1.075 = 10750
	total := int64(0)
	for _, inst := range result {
		total += inst.Amount
	}
	if total != 10750 {
		t.Errorf("expected total 10750, got %d", total)
	}

	for _, inst := range result {
		if inst.Status != string(domain.StatusPending) {
			t.Errorf("expected PENDING status, got %s", inst.Status)
		}
		if inst.DueDate == "" {
			t.Error("expected non-empty due date")
		}
	}
}

func TestCalculateInstallmentsRemainder(t *testing.T) {
	// 10000 / 3 = 3333.33... -> 3333 * 3 = 9999, remainder 1
	result := domain.CalculateInstallments(10000, 3, 0)
	if len(result) != 3 {
		t.Fatalf("expected 3 installments, got %d", len(result))
	}

	total := int64(0)
	for _, inst := range result {
		total += inst.Amount
	}
	if total != 10000 {
		t.Errorf("expected total 10000, got %d", total)
	}

	// Last installment should absorb the remainder
	last := result[2]
	first := result[0]
	if last.Amount != first.Amount+1 {
		t.Errorf("expected last installment (%d) to be first+1 (%d)", last.Amount, first.Amount+1)
	}
}

func TestCalculateInstallmentsTwelve(t *testing.T) {
	result := domain.CalculateInstallments(120000, 12, 1.5)
	if len(result) != 12 {
		t.Fatalf("expected 12 installments, got %d", len(result))
	}

	total := int64(0)
	for _, inst := range result {
		total += inst.Amount
	}

	// Total = 120000 * (1 + 1.5/100 * 12) = 120000 * 1.18 = 141600
	if total != 141600 {
		t.Errorf("expected total 141600, got %d", total)
	}
}

func TestNowTimestamp(t *testing.T) {
	ts := domain.NowTimestamp()
	if len(ts) != 23 {
		t.Errorf("expected 23 chars, got %d: %s", len(ts), ts)
	}
}

func TestFutureTimestamp(t *testing.T) {
	ts := domain.FutureTimestamp(30)
	if len(ts) != 23 {
		t.Errorf("expected 23 chars, got %d: %s", len(ts), ts)
	}
}
