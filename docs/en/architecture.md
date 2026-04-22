# Architecture

MockPay follows a layered architecture with clear separation between HTTP handling, business logic, data access, and domain models.

## Project Structure

```
mockpay/
├── main.go                          # Entrypoint, wiring, routes
├── config/
│   └── config.go                    # Environment configuration
├── internal/
│   ├── domain/                      # Domain models and business rules
│   │   ├── response.go              # API response envelope
│   │   ├── billing.go               # Billing, products, installments
│   │   ├── customer.go              # Customer models
│   │   ├── coupon.go                # Coupon models
│   │   ├── pix.go                   # PIX charge models
│   │   └── webhook.go               # Webhook event models
│   ├── store/
│   │   └── memory.go                # SQLite-backed store
│   ├── service/                     # Business logic layer
│   │   ├── billing.go               # Billing + installments + recurring
│   │   ├── customer.go              # Customer management
│   │   ├── coupon.go                # Coupon management
│   │   ├── pix.go                   # PIX charges
│   │   └── webhook.go               # Webhook dispatch + retry
│   ├── handler/                     # HTTP handlers
│   │   ├── billing.go               # Billing endpoints
│   │   ├── customer.go              # Customer endpoints
│   │   ├── coupon.go                # Coupon endpoints
│   │   ├── pix.go                   # PIX endpoints
│   │   ├── checkout.go              # Checkout + dashboard HTML
│   │   └── ui/                      # Embedded HTML templates
│   │       ├── checkout.html
│   │       └── dashboard.html
│   ├── middleware/
│   │   └── auth.go                  # Bearer token authentication
│   └── util/
│       ├── id.go                    # Random ID generation
│       └── crypto.go                # HMAC-SHA256 signatures
├── tests/                           # Integration and unit tests
├── docs/                            # Documentation
├── .env.example                     # Environment variable template
└── go.mod
```

## Dependency Flow

```
Config
  └─→ MemoryStore (SQLite)
        ├─→ WebhookService
        │     └─→ (background goroutines)
        ├─→ BillingService (uses WebhookService)
        ├─→ CustomerService
        ├─→ CouponService
        └─→ PixService (uses WebhookService)
              │
              ├─→ Handlers (Billing, Customer, Coupon, Pix)
              └─→ CheckoutHandler (uses Store + Billing + Pix services)
                    │
                    └─→ Fiber Routes
```

## Layers

### Domain (`internal/domain/`)

Pure data models with no dependencies on frameworks or external packages. Contains:
- Entity structs (Billing, PixCharge, Customer, Coupon)
- Status enums and constants
- Business calculation functions (`CalculateInstallments`, `NowTimestamp`)
- Request/response DTOs

### Store (`internal/store/`)

Data persistence layer using SQLite via `modernc.org/sqlite` (pure Go, no CGO). Complex fields (arrays, maps) are serialized as JSON text columns. All methods are thread-safe via `sync.RWMutex`.

### Service (`internal/service/`)

Business logic layer. Each service encapsulates validation, state transitions, and side effects (webhooks). Services depend on the store and optionally on each other (e.g., BillingService depends on WebhookService).

### Handler (`internal/handler/`)

HTTP layer using Fiber v3. API handlers return JSON. The CheckoutHandler renders HTML templates using `go:embed` and `html/template`.

### Middleware (`internal/middleware/`)

`Auth` middleware validates `Authorization: Bearer <token>` against the configured API key. Applied to all `/v1/*` routes. Checkout, dashboard, and health endpoints are public.

## Background Jobs

A single goroutine with a 30-second ticker handles:

1. **PIX Expiration** - Marks pending PIX charges past their `expires_at` as `EXPIRED`, dispatching `pix.expired` webhook.
2. **Recurring Billing** - Scans `MULTIPLE_PAYMENTS` billings with a past `next_billing` date, auto-creating new PENDING billings from the same template and dispatching `billing.created` webhook.

## IDs

All entity IDs are 32-character hex strings generated from 16 cryptographically random bytes. No prefixes or sequential patterns.

## Monetary Values

All amounts are integers in **cents** (Brazilian Real). For example, `R$ 50,00` is represented as `5000`. This avoids floating-point precision issues.

## Response Format

Every API response follows a standard envelope:

**Success:**
```json
{
  "data": { ... },
  "error": null
}
```

**Error:**
```json
{
  "data": null,
  "error": {
    "message": "Error description",
    "code": "ERROR_CODE"
  }
}
```
