package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/wesleiramos/mockpay/internal/domain"
	"github.com/wesleiramos/mockpay/internal/service"
)

type CouponHandler struct {
	service *service.CouponService
}

func NewCouponHandler(s *service.CouponService) *CouponHandler {
	return &CouponHandler{service: s}
}

func (h *CouponHandler) Create(c fiber.Ctx) error {
	var req domain.CreateCouponRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(domain.Err("invalid request body", "BAD_REQUEST"))
	}

	coupon, err := h.service.Create(req)
	if err != nil {
		return c.Status(400).JSON(domain.Err(err.Error(), "BAD_REQUEST"))
	}

	return c.Status(201).JSON(domain.Ok(coupon))
}

func (h *CouponHandler) List(c fiber.Ctx) error {
	coupons := h.service.List()
	return c.JSON(domain.Ok(coupons))
}
