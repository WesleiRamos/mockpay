# MockPay Playground

Web interface to test the MockPay server.

## Quick Start

```bash
bun start
```

Open http://localhost:3000

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | Playground port |
| `MOCKPAY_URL` | `http://localhost:8080` | MockPay server URL |
| `MOCKPAY_KEY` | `mock_key` | API key |

On the MockPay server, set:
```
MOCKPAY_WEBHOOK_URL=http://localhost:3000/webhook
```

## Features

- **Health** - Check server connection
- **Customers** - Create and list customers
- **Coupons** - Create and list discount coupons
- **Billings** - Create charges (PIX and/or card)
- **Cancel** - Cancel pending billings and stop recurring cycles
- **PIX** - Instant charges with QR Code
- **Stats** - Aggregated statistics

## Endpoints

- `GET /` - Main interface
- `GET /events` - Server-Sent Events for real-time status
- `POST /webhook` - Webhook receiver
