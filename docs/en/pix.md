# PIX

PIX charges represent instant payment requests. Each charge generates a QR code that can be scanned to approve the payment.

## Create a PIX Charge

```
POST /v1/pix/create
```

```json
{
  "amount": 5000,
  "expires_in": 3600,
  "description": "Payment description",
  "customer": {
    "name": "Joao Silva",
    "email": "joao@email.com",
    "cellphone": "11999999999",
    "tax_id": "12345678901"
  },
  "external_id": "charge-67890",
  "metadata": {
    "order_ref": "ORD-2026-001"
  }
}
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `amount` | int | yes | Amount in cents (must be >= 100) |
| `expires_in` | int | no | Expiration in seconds (default: 3600 = 1 hour) |
| `description` | string | no | Payment description |
| `customer` | object | no | Customer data (creates a new customer) |
| `external_id` | string | no | Your own identifier for this charge |
| `metadata` | object | no | Key-value pairs for additional data |

### Response

```json
{
  "data": {
    "id": "a1b2c3d4...",
    "amount": 5000,
    "status": "PENDING",
    "dev_mode": true,
    "br_code": "http://192.168.1.100:8080/checkout/a1b2c3d4.../approve",
    "br_code_base64": "data:image/png;base64,iVBOR...",
    "platform_fee": 80,
    "expires_at": "2026-04-21T13:00:00.000",
    "customer": {...},
    "external_id": "charge-67890",
    "metadata": {
      "order_ref": "ORD-2026-001"
    },
    "created_at": "2026-04-21T12:00:00.000",
    "updated_at": "2026-04-21T12:00:00.000"
  },
  "error": null
}
```

## Check PIX Status

```
GET /v1/pix/check?id=<id>
```

Returns the current PIX charge with its status.

## QR Code

Each PIX charge generates a real QR code (PNG image, 256x256px) containing the approve URL:

```
http://<hostname>:<port>/checkout/<id>/approve
```

The QR code is available in two formats:
- `br_code` - the plain URL string
- `br_code_base64` - a base64 data URI (`data:image/png;base64,...`) suitable for embedding in HTML `<img>` tags

### Hostname Configuration

The hostname in the QR URL is controlled by `MOCKPAY_HOSTNAME` (default: `localhost`). For mobile testing on your local network:

```bash
MOCKPAY_HOSTNAME=192.168.1.100 go run main.go
```

This generates QR codes containing `http://192.168.1.100:8080/checkout/.../approve`, which mobile devices can reach.

Scanning the QR code with any phone opens the approve endpoint directly, confirming the payment.

## Status Lifecycle

```
PENDING → APPROVED    (via checkout approve or QR scan)
PENDING → EXPIRED     (automatic, after expires_at passes)
```

## Expiration

PIX charges have an `expires_at` timestamp. A background job runs every 30 seconds to check for expired charges:

1. Lists all PIX charges with status `PENDING`
2. Parses each `expires_at`
3. If past expiration, sets status to `EXPIRED`
4. Dispatches `pix.expired` webhook

### Default Expiration

If `expires_in` is not provided, charges expire after **3600 seconds** (1 hour).

## Platform Fee

Each PIX charge includes a `platform_fee` field, hardcoded at **80 cents**. This is a simulated value for testing purposes.

## Approval Flow

1. PIX charge is created with `PENDING` status
2. The checkout page at `/checkout/<id>` displays the QR code
3. Approval can happen via:
   - Clicking "Approve payment" on the checkout page
   - Scanning the QR code (opens the approve URL)
   - Using the dashboard at `/`
4. On approval: status changes to `APPROVED`, `pix.approved` webhook fires
