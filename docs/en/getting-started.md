# Getting Started

MockPay is a payment gateway simulator for development and testing. It simulates Brazilian payment flows (PIX, credit card, installments, coupons) with a visual dashboard and checkout interface.

## Requirements

- Go 1.26+

## Install & Run

```bash
git clone git@github.com:WesleiRamos/mockpay.git && cd mockpay
go mod download
go run main.go
```

The server starts at `http://localhost:8080`.

## Configuration

Create a `.env` file from the example:

```bash
cp .env.example .env
```

| Variable | Default | Description |
|----------|---------|-------------|
| `MOCKPAY_PORT` | `8080` | HTTP listen port |
| `MOCKPAY_API_KEY` | `mock_key` | Bearer token for API authentication |
| `MOCKPAY_BASE_URL` | `http://localhost:<port>` | Base URL used to build checkout links |
| `MOCKPAY_PUBLIC_URL` | `MOCKPAY_BASE_URL` | Public URL used in QR codes (set to your LAN IP for mobile testing, falls back to BASE_URL) |
| `MOCKPAY_DB_PATH` | `mockpay.db` | SQLite database file path |
| `MOCKPAY_INTEREST_RATE` | `0` | Default monthly interest rate % for installments |
| `MOCKPAY_WEBHOOK_URL` | (empty) | URL to receive webhook events (empty = disabled) |
| `MOCKPAY_WEBHOOK_SECRET` | (empty) | HMAC-SHA256 secret for webhook signatures |

Environment variables are loaded via `.env` file (using `godotenv`) or from the system environment.

## Quick Test

```bash
# Create a billing
curl -X POST http://localhost:8080/v1/billing/create \
  -H "Authorization: Bearer mock_key" \
  -H "Content-Type: application/json" \
  -d '{
    "frequency": "ONE_TIME",
    "methods": ["PIX"],
    "products": [{"external_id": "p1", "name": "Test Product", "quantity": 1, "price": 5000}],
    "return_url": "http://localhost:3000/return",
    "completion_url": "http://localhost:3000/done"
  }'

# Open the dashboard
open http://localhost:8080/

# Create a PIX charge
curl -X POST http://localhost:8080/v1/pix/create \
  -H "Authorization: Bearer mock_key" \
  -H "Content-Type: application/json" \
  -d '{"amount": 5000, "expires_in": 3600}'

# Check health
curl http://localhost:8080/health
```

## Mobile Testing

To test QR codes from a mobile device on the same network:

```bash
MOCKPAY_PUBLIC_URL=http://192.168.1.100:8080 go run main.go
```

The QR codes will contain URLs using your LAN IP instead of `localhost`, allowing mobile devices to reach the server.

## Building

```bash
go build -o mockpay .
./mockpay
```

## Running Tests

```bash
go test ./...
```

Tests use in-memory SQLite (`:memory:`) for isolation.
