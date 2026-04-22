package service

import (
	"fmt"

	"github.com/wesleiramos/mockpay/internal/domain"
	"github.com/wesleiramos/mockpay/internal/store"
	"github.com/wesleiramos/mockpay/internal/util"
)

type CouponService struct {
	store *store.MemoryStore
}

func NewCouponService(s *store.MemoryStore) *CouponService {
	return &CouponService{store: s}
}

func (cs *CouponService) Create(req domain.CreateCouponRequest) (*domain.Coupon, error) {
	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}
	if req.Discount <= 0 {
		return nil, fmt.Errorf("discount must be > 0")
	}

	now := domain.NowTimestamp()
	coupon := &domain.Coupon{
		ID:           util.NewID(),
		Code:         req.Code,
		Notes:        req.Notes,
		MaxRedeems:   req.MaxRedeems,
		RedeemsCount: 0,
		DiscountKind: req.DiscountKind,
		Discount:     req.Discount,
		DevMode:      true,
		Status:       domain.CouponActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	cs.store.CreateCoupon(coupon)
	return coupon, nil
}

func (cs *CouponService) List() []*domain.Coupon {
	return cs.store.ListCoupons()
}

func (cs *CouponService) ApplyDiscount(couponID string, amount int64) (int64, error) {
	c, ok := cs.store.GetCoupon(couponID)
	if !ok {
		return 0, fmt.Errorf("coupon not found")
	}
	if c.Status != domain.CouponActive {
		return 0, fmt.Errorf("coupon is not active")
	}
	if c.MaxRedeems > 0 && c.RedeemsCount >= c.MaxRedeems {
		return 0, fmt.Errorf("coupon has no remaining uses")
	}

	var discounted int64
	switch c.DiscountKind {
	case domain.DiscountPercentage:
		discounted = amount * c.Discount / 100
	case domain.DiscountFixed:
		discounted = c.Discount
	}

	if discounted > amount {
		discounted = amount
	}

	cs.store.IncrementCouponRedeems(couponID)
	return amount - discounted, nil
}

func (cs *CouponService) ApplyDiscountByCode(code string, amount int64) (int64, error) {
	c, ok := cs.store.GetCouponByCode(code)
	if !ok {
		return 0, fmt.Errorf("coupon not found")
	}
	return cs.ApplyDiscount(c.ID, amount)
}
