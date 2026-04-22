package handler

import (
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/skip2/go-qrcode"
	"github.com/wesleiramos/mockpay/internal/domain"
	"github.com/wesleiramos/mockpay/internal/i18n"
	"github.com/wesleiramos/mockpay/internal/service"
	"github.com/wesleiramos/mockpay/internal/store"
)

//go:embed all:ui
var uiFS embed.FS

type CheckoutHandler struct {
	store          *store.MemoryStore
	billingService *service.BillingService
	pixService     *service.PixService
	baseURL        string
	templates      *template.Template
}

func NewCheckoutHandler(s *store.MemoryStore, bs *service.BillingService, ps *service.PixService, baseURL string) *CheckoutHandler {
	funcMap := template.FuncMap{
		"lower":   strings.ToLower,
		"printf":  fmt.Sprintf,
		"not":     func(b bool) bool { return !b },
		"iterate": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i + 1
			}
			return s
		},
		"t": func(key string) string { return key },
	}
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFS(uiFS, "ui/checkout.html", "ui/dashboard.html"))

	return &CheckoutHandler{
		store:          s,
		billingService: bs,
		pixService:     ps,
		baseURL:        baseURL,
		templates:      tmpl,
	}
}

type checkoutData struct {
	ID              string
	Amount          string
	OriginalAmount  string
	CouponCode      string
	RawAmount       int64
	Products        []productDisplay
	Status          string
	Method          string
	Methods         []string
	IsPIX           bool
	HasCard         bool
	IsBilling       bool
	CompletionURL   string
	ReturnURL       string
	Installments    []installmentDisplay
	MaxInstallments int
	InterestRate    float64
	QRCodeBase64    template.URL
	ExpiresAt       string
	Lang            string
}

type productDisplay struct {
	Name     string
	Quantity int
	Price    string
}

type installmentDisplay struct {
	Number  int
	Amount  string
	Status  string
	DueDate string
}

func formatAmount(cents int64) string {
	reais := cents / 100
	centavos := cents % 100
	return fmt.Sprintf("R$ %d,%02d", reais, centavos)
}

func getLang(c fiber.Ctx) string {
	lang := c.Cookies("mockpay_lang", "pt-BR")
	return i18n.ValidLang(lang)
}

func (h *CheckoutHandler) render(c fiber.Ctx, name string, data any) error {
	lang := getLang(c)
	funcs := template.FuncMap{
		"t": i18n.GetFunc(lang),
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	return h.templates.Funcs(funcs).ExecuteTemplate(c, name, data)
}

func generateQRBase64(content string) template.URL {
	png, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		log.Printf("qr code error: %v", err)
		return ""
	}
	return template.URL("data:image/png;base64," + base64.StdEncoding.EncodeToString(png))
}

func (h *CheckoutHandler) CheckoutPage(c fiber.Ctx) error {
	id := c.Params("id")

	item, found := h.store.FindByID(id)
	if !found {
		return c.Status(404).SendString("Payment not found")
	}

	data := checkoutData{
		ID:              id,
		MaxInstallments: 1,
		InterestRate:    0,
		Lang:            getLang(c),
	}

	approveURL := fmt.Sprintf("%s/checkout/%s/approve", h.baseURL, id)

	switch v := item.(type) {
	case *domain.Billing:
		data.IsBilling = true
		data.Amount = formatAmount(v.Amount)
		data.RawAmount = v.Amount
		if v.OriginalAmount > 0 {
			data.OriginalAmount = formatAmount(v.OriginalAmount)
			data.CouponCode = v.CouponCode
		}
		data.Status = string(v.Status)
		data.CompletionURL = v.CompletionURL
		data.ReturnURL = v.ReturnURL
		data.MaxInstallments = v.Installments
		data.InterestRate = v.InterestRate
		if data.MaxInstallments < 1 {
			data.MaxInstallments = 1
		}
		for _, m := range v.Methods {
			data.Methods = append(data.Methods, string(m))
			if m == domain.MethodPIX {
				data.IsPIX = true
			}
			if m == domain.MethodCARD {
				data.HasCard = true
			}
		}
		if len(data.Methods) > 0 {
			data.Method = data.Methods[0]
		}
		for _, p := range v.Products {
			data.Products = append(data.Products, productDisplay{
				Name:     p.Name,
				Quantity: p.Quantity,
				Price:    formatAmount(p.Price),
			})
		}
		for _, inst := range v.InstallmentList {
			data.Installments = append(data.Installments, installmentDisplay{
				Number:  inst.Number,
				Amount:  formatAmount(inst.Amount),
				Status:  inst.Status,
				DueDate: inst.DueDate,
			})
		}
		if v.Status == domain.StatusPending {
			data.QRCodeBase64 = generateQRBase64(approveURL)
		}
	case *domain.PixCharge:
		data.IsPIX = true
		data.Amount = formatAmount(v.Amount)
		data.Status = string(v.Status)
		data.Method = "PIX"
		data.ExpiresAt = v.ExpiresAt
		if v.Status == domain.PixPending {
			data.QRCodeBase64 = generateQRBase64(approveURL)
		}
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return h.render(c, "checkout.html", data)
}

func (h *CheckoutHandler) Approve(c fiber.Ctx) error {
	id := c.Params("id")
	var redirectURL string

	if b, ok := h.store.GetBilling(id); ok {
		_, err := h.billingService.Approve(id)
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}
		redirectURL = b.CompletionURL
	} else if _, ok := h.store.GetPixCharge(id); ok {
		_, err := h.pixService.Approve(id)
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}
	}

	if redirectURL != "" {
		return c.Redirect().To(redirectURL)
	}
	return c.Redirect().To("/checkout/" + id)
}

func (h *CheckoutHandler) Deny(c fiber.Ctx) error {
	id := c.Params("id")
	var redirectURL string

	if b, ok := h.store.GetBilling(id); ok {
		_, err := h.billingService.Deny(id)
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}
		redirectURL = b.ReturnURL
	} else if _, ok := h.store.GetPixCharge(id); ok {
		_, err := h.pixService.Deny(id)
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}
	}

	if redirectURL != "" {
		return c.Redirect().To(redirectURL)
	}
	return c.Redirect().To("/checkout/" + id)
}

func (h *CheckoutHandler) Dashboard(c fiber.Ctx) error {
	entries := h.store.ListAllPayments()

	type dashboardPayment struct {
		ID        string
		Amount    string
		Status    string
		Method    string
		Type      string
		CreatedAt string
	}

	payments := make([]dashboardPayment, len(entries))
	for i, e := range entries {
		payments[i] = dashboardPayment{
			ID:        e.ID,
			Amount:    formatAmount(e.Amount),
			Status:    e.Status,
			Method:    e.Method,
			Type:      e.Type,
			CreatedAt: e.CreatedAt,
		}
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return h.render(c, "dashboard.html", map[string]any{
		"Payments": payments,
		"Lang":     getLang(c),
	})
}

func (h *CheckoutHandler) Stats(c fiber.Ctx) error {
	return c.JSON(domain.Ok(h.store.Stats()))
}

func (h *CheckoutHandler) Health(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func StartBackgroundJobs(bs *service.BillingService, ps *service.PixService) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("[background] checking expired pix charges...")
			ps.CheckExpired()

			log.Println("[background] checking recurring billings...")
			bs.CheckRecurring()
		}
	}()
}
