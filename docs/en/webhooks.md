# Webhooks

MockPay can send real HTTP webhook notifications when payment events occur. Webhooks are disabled by default and must be configured via environment variables.

## Configuration

```bash
MOCKPAY_WEBHOOK_URL=https://your-server.com/webhooks
MOCKPAY_WEBHOOK_SECRET=your-secret-key
```

Both variables are optional. If `MOCKPAY_WEBHOOK_URL` is empty, webhooks are disabled.

## Events

| Event | Trigger |
|-------|---------|
| `billing.approved` | Billing payment approved |
| `billing.denied` | Billing payment denied |
| `billing.cancelled` | Billing cancelled via API |
| `billing.created` | Recurring billing auto-created |
| `pix.approved` | PIX payment approved |
| `pix.expired` | PIX charge expired |

## Payload

Each webhook request is a `POST` with `Content-Type: application/json`:

```json
{
  "id": "evt_xxxx",
  "type": "billing.approved",
  "payload": {
    "id": "abc123...",
    "amount": 5000,
    "status": "APPROVED",
    ...
  },
  "created_at": "2026-04-21T12:00:00.000"
}
```

The `payload` contains the full entity (billing or PIX charge) at the time of the event.

## Signature Verification

Each webhook request includes an `X-MockPay-Signature` header:

```
X-MockPay-Signature: sha256=<hex_digest>
```

The signature is computed as **HMAC-SHA256** of the raw request body using the webhook secret.

### Verifying in Your Application

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "strings"
)

func verifySignature(secret string, body []byte, header string) bool {
    if !strings.HasPrefix(header, "sha256=") {
        return false
    }
    sig, _ := hex.DecodeString(header[7:])

    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(body)
    expected := mac.Sum(nil)

    return hmac.Equal(sig, expected)
}
```

### Verifying with curl

```bash
# Test your webhook receiver locally
# Using webhook.site or similar services for testing

MOCKPAY_WEBHOOK_URL=https://webhook.site/your-unique-url
MOCKPAY_WEBHOOK_SECRET=my-secret
```

## Delivery & Retry

Webhooks are sent asynchronously via a buffered channel (capacity 100):

1. **Initial send** - HTTP POST with 5-second timeout
2. **On failure** (non-2xx response or network error):
   - Record the delivery attempt in the database
   - Enqueue for retry
3. **Retries** - Up to 3 total attempts with linear backoff:
   - Attempt 1: immediate
   - Attempt 2: after 5 seconds
   - Attempt 3: after 10 seconds

### Delivery Records

Each delivery attempt is stored in the `webhook_deliveries` table with:
- `id` - unique delivery ID
- `event_id` - the event being delivered
- `url` - the target URL
- `attempt` - attempt number (1-3)
- `status_code` - HTTP response code (0 on network error)
- `success` - whether the delivery succeeded (2xx response)

## Architecture

The webhook system uses two background goroutines:

1. **Worker** - reads from the events channel and sends HTTP requests
2. **Retry Worker** - reads from the retries channel, applies backoff delay, then resends

```
Dispatch() → events channel → worker → HTTP POST
                                         ↓ failure
                                    retries channel → retryWorker → HTTP POST
```

This design ensures webhook delivery doesn't block the main request handling.
