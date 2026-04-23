package domain

type PixStatus string

const (
	PixPending  PixStatus = "PENDING"
	PixApproved PixStatus = "APPROVED"
	PixExpired  PixStatus = "EXPIRED"
)

type PixCharge struct {
	ID           string            `json:"id"`
	Amount       int64             `json:"amount"`
	Status       PixStatus         `json:"status"`
	DevMode      bool              `json:"dev_mode"`
	BrCode       string            `json:"br_code"`
	BrCodeBase64 string            `json:"br_code_base64"`
	PlatformFee  int64             `json:"platform_fee"`
	ExpiresAt    string            `json:"expires_at"`
	Customer     *CustomerRef      `json:"customer,omitempty"`
	ExternalID   string            `json:"external_id,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    string            `json:"created_at"`
	UpdatedAt    string            `json:"updated_at"`
}

type CreatePixRequest struct {
	Amount      int64             `json:"amount"`
	ExpiresIn   int               `json:"expires_in,omitempty"`
	Description string            `json:"description,omitempty"`
	Customer    *CustomerInput    `json:"customer,omitempty"`
	ExternalID  string            `json:"external_id,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}
