package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/wesleiramos/mockpay/config"
	"github.com/wesleiramos/mockpay/internal/handler"
	"github.com/wesleiramos/mockpay/internal/middleware"
	"github.com/wesleiramos/mockpay/internal/service"
	"github.com/wesleiramos/mockpay/internal/store"
)

func main() {
	cfg := config.Load()

	memStore := store.NewWithPath(cfg.DBPath)
	defer memStore.Close()
	webhookSvc := service.NewWebhookService(memStore, cfg.WebhookURL, cfg.WebhookSecret)
	couponSvc := service.NewCouponService(memStore)
	billingSvc := service.NewBillingService(memStore, cfg.BaseURL, webhookSvc, couponSvc, cfg.DefaultInterestRate)
	customerSvc := service.NewCustomerService(memStore)

	qrBaseURL := cfg.PublicURL
	pixSvc := service.NewPixService(memStore, webhookSvc, qrBaseURL)

	billingH := handler.NewBillingHandler(billingSvc)
	customerH := handler.NewCustomerHandler(customerSvc)
	couponH := handler.NewCouponHandler(couponSvc)
	pixH := handler.NewPixHandler(pixSvc)
	checkoutH := handler.NewCheckoutHandler(memStore, billingSvc, pixSvc, qrBaseURL)

	handler.StartBackgroundJobs(billingSvc, pixSvc)

	app := fiber.New(fiber.Config{
		AppName: "MockPay",
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}))

	api := app.Group("/v1", middleware.Auth(cfg))
	api.Post("/billing/create", billingH.Create)
	api.Get("/billing/get", billingH.Get)
	api.Get("/billing/list", billingH.List)
	api.Get("/billing/:id/installments", billingH.GetInstallments)
	api.Post("/billing/:id/cancel", billingH.Cancel)

	api.Post("/pix/create", pixH.Create)
	api.Get("/pix/check", pixH.Check)

	api.Post("/customer/create", customerH.Create)
	api.Get("/customer/list", customerH.List)

	api.Post("/coupon/create", couponH.Create)
	api.Get("/coupon/list", couponH.List)

	api.Get("/stats", checkoutH.Stats)

	app.Get("/checkout/:id", checkoutH.CheckoutPage)
	app.Get("/checkout/:id/approve", checkoutH.Approve)
	app.Get("/checkout/:id/deny", checkoutH.Deny)
	app.Get("/health", checkoutH.Health)
	app.Get("/", checkoutH.Dashboard)

	log.Printf("MockPay starting on :%s", cfg.Port)
	log.Printf("Dashboard: %s/", cfg.PublicURL)
	log.Printf("Public URL: %s (set MOCKPAY_PUBLIC_URL to change)", cfg.PublicURL)
	log.Printf("API Key: %s", cfg.APIKey)
	log.Printf("Webhook URL: %s", cfg.WebhookURL)

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
