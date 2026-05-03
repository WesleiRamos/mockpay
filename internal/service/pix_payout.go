package service

import (
	"fmt"
	"log"
	"time"

	"github.com/wesleiramos/mockpay/internal/domain"
	"github.com/wesleiramos/mockpay/internal/store"
	"github.com/wesleiramos/mockpay/internal/util"
)

type PixPayoutService struct {
	store   *store.MemoryStore
	webhook *WebhookService
}

func NewPixPayoutService(s *store.MemoryStore, w *WebhookService) *PixPayoutService {
	return &PixPayoutService{store: s, webhook: w}
}

func (ps *PixPayoutService) CreatePayout(req domain.CreatePixPayoutRequest) (*domain.PixPayout, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be > 0")
	}

	validKeyTypes := map[string]bool{"cpf": true, "phone": true, "email": true, "random": true}
	if !validKeyTypes[req.PixKeyType] {
		return nil, fmt.Errorf("invalid pix_key_type: must be cpf, phone, email, or random")
	}

	if req.PixKey == "" {
		return nil, fmt.Errorf("pix_key is required")
	}

	if req.IdempotencyKey == "" {
		return nil, fmt.Errorf("idempotency_key is required")
	}

	existing, ok := ps.store.GetPixPayoutByIdempotencyKey(req.IdempotencyKey)
	if ok {
		return existing, nil
	}

	now := domain.NowTimestamp()

	payout := &domain.PixPayout{
		ID:             util.NewID(),
		Amount:         req.Amount,
		Status:         domain.PixPayoutProcessing,
		ExternalID:     req.ExternalID,
		EndToEndID:     "",
		PixKeyType:     req.PixKeyType,
		PixKey:         req.PixKey,
		IdempotencyKey: req.IdempotencyKey,
		Metadata:       req.Metadata,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	ps.store.CreatePixPayout(payout)
	return payout, nil
}

func (ps *PixPayoutService) GetPayout(id string) (*domain.PixPayout, error) {
	p, ok := ps.store.GetPixPayout(id)
	if !ok {
		return nil, fmt.Errorf("pix payout not found")
	}
	return p, nil
}

func (ps *PixPayoutService) CheckLiquidate() {
	payouts := ps.store.ListPixPayouts()
	for _, p := range payouts {
		if p.Status != domain.PixPayoutProcessing {
			continue
		}

		createdAt, err := time.Parse("2006-01-02T15:04:05.000", p.CreatedAt)
		if err != nil {
			continue
		}

		if time.Since(createdAt) < 30*time.Second {
			continue
		}

		endToEndID := generateEndToEndID(p.ID)
		ps.store.UpdatePixPayoutStatus(p.ID, domain.PixPayoutLiquidated, endToEndID)
		updated, _ := ps.store.GetPixPayout(p.ID)
		ps.webhook.Dispatch(domain.EventPixPayoutLiquidated, updated)
		log.Printf("pix payout %s liquidated", p.ID)
	}
}

func generateEndToEndID(payoutID string) string {
	ts := time.Now().UTC().Format("20060102")
	short := payoutID
	if len(short) > 34 {
		short = short[:34]
	}
	for len(short) < 34 {
		short += "0"
	}
	return "E" + ts + short[0:8] + short[8:18] + short[18:28] + short[28:34]
}
