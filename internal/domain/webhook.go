package domain

type WebhookEventType string

const (
	EventBillingApproved WebhookEventType = "billing.approved"
	EventBillingDenied   WebhookEventType = "billing.denied"
	EventBillingCreated   WebhookEventType = "billing.created"
	EventBillingCancelled WebhookEventType = "billing.cancelled"
	EventPixApproved      WebhookEventType = "pix.approved"
	EventPixExpired      WebhookEventType = "pix.expired"
)

type WebhookEvent struct {
	ID         string           `json:"id"`
	Type       WebhookEventType `json:"type"`
	Payload    any              `json:"payload"`
	CreatedAt  string           `json:"created_at"`
}

type WebhookDelivery struct {
	ID         string `json:"id"`
	EventID    string `json:"event_id"`
	URL        string `json:"url"`
	Attempt    int    `json:"attempt"`
	StatusCode int    `json:"status_code"`
	Success    bool   `json:"success"`
	CreatedAt  string `json:"created_at"`
}
