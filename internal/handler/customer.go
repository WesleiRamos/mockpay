package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/wesleiramos/mockpay/internal/domain"
	"github.com/wesleiramos/mockpay/internal/service"
)

type CustomerHandler struct {
	service *service.CustomerService
}

func NewCustomerHandler(s *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: s}
}

func (h *CustomerHandler) Create(c fiber.Ctx) error {
	var req domain.CustomerInput
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(domain.Err("invalid request body", "BAD_REQUEST"))
	}

	customer, err := h.service.Create(req)
	if err != nil {
		return c.Status(400).JSON(domain.Err(err.Error(), "BAD_REQUEST"))
	}

	return c.Status(201).JSON(domain.Ok(customer))
}

func (h *CustomerHandler) List(c fiber.Ctx) error {
	customers := h.service.List()
	return c.JSON(domain.Ok(customers))
}
