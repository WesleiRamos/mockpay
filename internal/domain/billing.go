package domain

import "time"

type BillingStatus string

const (
	StatusPending   BillingStatus = "PENDING"
	StatusApproved  BillingStatus = "APPROVED"
	StatusDenied    BillingStatus = "DENIED"
	StatusCancelled BillingStatus = "CANCELLED"
	StatusExpired   BillingStatus = "EXPIRED"
)

type BillingFrequency string

const (
	FrequencyOneTime        BillingFrequency = "ONE_TIME"
	FrequencyMultipleBilling BillingFrequency = "MULTIPLE_PAYMENTS"
)

type BillingMethod string

const (
	MethodPIX  BillingMethod = "PIX"
	MethodCARD BillingMethod = "CARD"
)

type Billing struct {
	ID              string          `json:"id"`
	URL             string          `json:"url"`
	Amount          int64           `json:"amount"`
	OriginalAmount  int64           `json:"original_amount,omitempty"`
	CouponCode      string          `json:"coupon_code,omitempty"`
	Status          BillingStatus   `json:"status"`
	DevMode         bool            `json:"dev_mode"`
	Methods         []BillingMethod `json:"methods"`
	Products        []Product       `json:"products"`
	Frequency       BillingFrequency `json:"frequency"`
	NextBilling     *string         `json:"next_billing"`
	Customer        *CustomerRef    `json:"customer,omitempty"`
	ReturnURL       string          `json:"return_url,omitempty"`
	CompletionURL   string          `json:"completion_url,omitempty"`
	Installments    int             `json:"installments"`
	InterestRate    float64         `json:"interest_rate"`
	InstallmentList []Installment   `json:"installment_list,omitempty"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}

type Installment struct {
	Number  int     `json:"number"`
	Amount  int64   `json:"amount"`
	Status  string  `json:"status"`
	DueDate string  `json:"due_date"`
	PaidAt  *string `json:"paid_at,omitempty"`
}

type Product struct {
	ExternalID  string `json:"external_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Quantity    int    `json:"quantity"`
	Price       int64  `json:"price"`
}

type CreateBillingRequest struct {
	Frequency     BillingFrequency `json:"frequency"`
	Methods       []BillingMethod  `json:"methods"`
	Products      []Product        `json:"products"`
	ReturnURL     string           `json:"return_url"`
	CompletionURL string           `json:"completion_url"`
	CustomerID    string           `json:"customer_id,omitempty"`
	Customer      *CustomerInput   `json:"customer,omitempty"`
	Installments  int              `json:"installments,omitempty"`
	InterestRate  float64          `json:"interest_rate,omitempty"`
	CouponCode    string           `json:"coupon_code,omitempty"`
}

func NowTimestamp() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000")
}

func FutureTimestamp(days int) string {
	return time.Now().UTC().AddDate(0, 0, days).Format("2006-01-02T15:04:05.000")
}

func CalculateInstallments(totalAmount int64, count int, rate float64) []Installment {
	if count <= 1 {
		return nil
	}

	totalWithInterest := float64(totalAmount) * (1 + rate/100*float64(count))
	totalCents := int64(totalWithInterest)

	perInstallment := totalCents / int64(count)
	remainder := totalCents % int64(count)

	now := time.Now().UTC()
	installments := make([]Installment, count)

	for i := 0; i < count; i++ {
		amt := perInstallment
		if i == count-1 {
			amt += remainder
		}
		dueDate := now.AddDate(0, 0, 30*(i+1)).Format("2006-01-02T15:04:05.000")
		installments[i] = Installment{
			Number:  i + 1,
			Amount:  amt,
			Status:  string(StatusPending),
			DueDate: dueDate,
		}
	}

	return installments
}
