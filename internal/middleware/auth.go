package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/wesleiramos/mockpay/config"
)

func Auth(cfg *config.Config) fiber.Handler {
	return func(c fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(401).JSON(fiber.Map{
				"data":  nil,
				"error": fiber.Map{"message": "Authorization header is required", "code": "UNAUTHORIZED"},
			})
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		if token == auth || token != cfg.APIKey {
			return c.Status(401).JSON(fiber.Map{
				"data":  nil,
				"error": fiber.Map{"message": "Invalid API key", "code": "UNAUTHORIZED"},
			})
		}

		return c.Next()
	}
}
