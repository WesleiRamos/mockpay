package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/wesleiramos/mockpay/internal/domain"
	"github.com/wesleiramos/mockpay/internal/service"
)

type PixPayoutHandler struct {
	service *service.PixPayoutService
}

func NewPixPayoutHandler(s *service.PixPayoutService) *PixPayoutHandler {
	return &PixPayoutHandler{service: s}
}

func (h *PixPayoutHandler) Create(c fiber.Ctx) error {
	var req domain.CreatePixPayoutRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(domain.Err("invalid request body", "BAD_REQUEST"))
	}

	payout, err := h.service.CreatePayout(req)
	if err != nil {
		return c.Status(400).JSON(domain.Err(err.Error(), "BAD_REQUEST"))
	}

	return c.Status(201).JSON(domain.Ok(payout))
}

func (h *PixPayoutHandler) Check(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(domain.Err("id parameter is required", "BAD_REQUEST"))
	}

	payout, err := h.service.GetPayout(id)
	if err != nil {
		return c.Status(404).JSON(domain.Err(err.Error(), "NOT_FOUND"))
	}

	return c.JSON(domain.Ok(payout))
}
