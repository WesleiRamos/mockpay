package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/wesleiramos/mockpay/internal/domain"
	"github.com/wesleiramos/mockpay/internal/service"
)

type PixHandler struct {
	service *service.PixService
}

func NewPixHandler(s *service.PixService) *PixHandler {
	return &PixHandler{service: s}
}

func (h *PixHandler) Create(c fiber.Ctx) error {
	var req domain.CreatePixRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(domain.Err("invalid request body", "BAD_REQUEST"))
	}

	charge, err := h.service.Create(req)
	if err != nil {
		return c.Status(400).JSON(domain.Err(err.Error(), "BAD_REQUEST"))
	}

	return c.Status(201).JSON(domain.Ok(charge))
}

func (h *PixHandler) Check(c fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return c.Status(400).JSON(domain.Err("id parameter is required", "BAD_REQUEST"))
	}

	charge, err := h.service.Get(id)
	if err != nil {
		return c.Status(404).JSON(domain.Err(err.Error(), "NOT_FOUND"))
	}

	return c.JSON(domain.Ok(charge))
}
