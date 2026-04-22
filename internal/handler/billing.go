package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/wesleiramos/mockpay/internal/domain"
	"github.com/wesleiramos/mockpay/internal/service"
)

type BillingHandler struct {
	service *service.BillingService
}

func NewBillingHandler(s *service.BillingService) *BillingHandler {
	return &BillingHandler{service: s}
}

func (h *BillingHandler) Create(c fiber.Ctx) error {
	var req domain.CreateBillingRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(domain.Err("invalid request body", "BAD_REQUEST"))
	}

	billing, err := h.service.Create(req)
	if err != nil {
		return c.Status(400).JSON(domain.Err(err.Error(), "BAD_REQUEST"))
	}

	return c.Status(201).JSON(domain.Ok(billing))
}

func (h *BillingHandler) Get(c fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return c.Status(400).JSON(domain.Err("id parameter is required", "BAD_REQUEST"))
	}

	billing, err := h.service.Get(id)
	if err != nil {
		return c.Status(404).JSON(domain.Err(err.Error(), "NOT_FOUND"))
	}

	return c.JSON(domain.Ok(billing))
}

func (h *BillingHandler) List(c fiber.Ctx) error {
	billings := h.service.List()
	return c.JSON(domain.Ok(billings))
}

func (h *BillingHandler) GetInstallments(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(domain.Err("id parameter is required", "BAD_REQUEST"))
	}

	installments, err := h.service.GetInstallments(id)
	if err != nil {
		return c.Status(404).JSON(domain.Err(err.Error(), "NOT_FOUND"))
	}

	return c.JSON(domain.Ok(installments))
}
