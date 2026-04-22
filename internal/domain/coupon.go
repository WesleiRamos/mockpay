package domain

type CouponStatus string

const (
	CouponActive   CouponStatus = "ACTIVE"
	CouponInactive CouponStatus = "INACTIVE"
)

type DiscountKind string

const (
	DiscountPercentage DiscountKind = "PERCENTAGE"
	DiscountFixed     DiscountKind = "FIXED"
)

type Coupon struct {
	ID           string       `json:"id"`
	Code         string       `json:"code"`
	Notes        string       `json:"notes,omitempty"`
	MaxRedeems   int          `json:"max_redeems"`
	RedeemsCount int          `json:"redeems_count"`
	DiscountKind DiscountKind `json:"discount_kind"`
	Discount     int64        `json:"discount"`
	DevMode      bool         `json:"dev_mode"`
	Status       CouponStatus `json:"status"`
	CreatedAt    string       `json:"created_at"`
	UpdatedAt    string       `json:"updated_at"`
}

type CreateCouponRequest struct {
	Code         string       `json:"code"`
	Notes        string       `json:"notes,omitempty"`
	MaxRedeems   int          `json:"max_redeems"`
	DiscountKind DiscountKind `json:"discount_kind"`
	Discount     int64        `json:"discount"`
}
