package domain

type PixPayoutStatus string

const (
	PixPayoutProcessing  PixPayoutStatus = "PROCESSING"
	PixPayoutLiquidated  PixPayoutStatus = "LIQUIDATED"
	PixPayoutFailed      PixPayoutStatus = "FAILED"
)

type PixPayout struct {
	ID              string            `json:"id"`
	Amount          int64             `json:"amount"`
	Status          PixPayoutStatus   `json:"status"`
	ExternalID      string            `json:"external_id"`
	EndToEndID      string            `json:"end_to_end_id"`
	PixKeyType      string            `json:"pix_key_type,omitempty"`
	PixKey          string            `json:"pix_key,omitempty"`
	IdempotencyKey  string            `json:"idempotency_key,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	CreatedAt       string            `json:"created_at"`
	UpdatedAt       string            `json:"updated_at"`
}

type CreatePixPayoutRequest struct {
	Amount         int64             `json:"amount"`
	PixKeyType     string            `json:"pix_key_type"`
	PixKey         string            `json:"pix_key"`
	ExternalID     string            `json:"external_id"`
	IdempotencyKey string            `json:"idempotency_key"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}
