# Billing

Billings are the primary payment entity. They represent a charge to a customer that can be paid via PIX or credit card, optionally split into installments with interest, and can be set to recur automatically.

## Create a Billing

```
POST /v1/billing/create
```

```json
{
  "frequency": "ONE_TIME",
  "methods": ["PIX", "CARD"],
  "products": [
    {
      "external_id": "prod-001",
      "name": "Product Name",
      "description": "Product description",
      "quantity": 1,
      "price": 5000
    }
  ],
  "return_url": "https://example.com/return",
  "completion_url": "https://example.com/done",
  "customer_id": "cust_xxxx",
  "customer": {
    "name": "Joao Silva",
    "email": "joao@email.com",
    "cellphone": "11999999999",
    "tax_id": "12345678901"
  },
  "installments": 3,
  "interest_rate": 2.5,
  "coupon_code": "DESCONTO20",
  "external_id": "order-12345",
  "metadata": {
    "source": "website",
    "channel": "organic"
  }
}
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `frequency` | string | yes | `ONE_TIME` or `MULTIPLE_PAYMENTS` |
| `methods` | string[] | yes | Payment methods: `PIX`, `CARD`, or both |
| `products` | object[] | yes | At least one product |
| `products[].external_id` | string | yes | Your product identifier |
| `products[].name` | string | yes | Product name |
| `products[].description` | string | no | Product description |
| `products[].quantity` | int | yes | Quantity (must be >= 1) |
| `products[].price` | int | yes | Price in cents (must be >= 100) |
| `return_url` | string | no | URL for "back" action after denial |
| `completion_url` | string | no | URL redirect after approval |
| `customer_id` | string | no | Existing customer ID |
| `customer` | object | no | Customer data (creates new customer) |
| `installments` | int | no | Number of installments (1-12, default 1) |
| `interest_rate` | float | no | Monthly interest rate % (default: env `MOCKPAY_INTEREST_RATE`) |
| `coupon_code` | string | no | Coupon code to apply a discount to the total amount |
| `external_id` | string | no | Your own identifier for this billing |
| `metadata` | object | no | Key-value pairs for additional data |

The total billing `amount` is the sum of all products (`price * quantity`).

### Response

```json
{
  "data": {
    "id": "a1b2c3d4...",
    "url": "http://localhost:8080/checkout/a1b2c3d4...",
    "amount": 5000,
    "original_amount": 5000,
    "coupon_code": "DESCONTO20",
    "status": "PENDING",
    "dev_mode": true,
    "methods": ["PIX", "CARD"],
    "products": [...],
    "frequency": "ONE_TIME",
    "next_billing": null,
    "installments": 3,
    "interest_rate": 2.5,
    "installment_list": [...],
    "external_id": "order-12345",
    "metadata": {
      "source": "website",
      "channel": "organic"
    },
    "created_at": "2026-04-21T12:00:00.000",
    "updated_at": "2026-04-21T12:00:00.000"
  },
  "error": null
}
```

## Get a Billing

```
GET /v1/billing/:id
```

## List Billings

```
GET /v1/billing/list
```

Returns all billings ordered by creation date (newest first).

## Status Lifecycle

```
PENDING → APPROVED
PENDING → DENIED
PENDING → EXPIRED (recurring only)
PENDING → CANCELLED
```

- **APPROVED** - Payment was approved via checkout or dashboard.
- **DENIED** - Payment was denied via checkout or dashboard.
- **Status transitions are final** - a billing cannot change status after being approved or denied.

## Approval Flow

1. Billing is created with `PENDING` status
2. A checkout page is available at `/checkout/<id>` with Approve/Deny buttons
3. On approve: status changes to `APPROVED`, all installments are marked as approved, `completion_url` webhook fires
4. On deny: status changes to `DENIED`, all installments are marked as denied, `return_url` is used for redirect

## Coupon / Discount

A coupon can be applied at billing creation time by passing `coupon_code`:

```json
{
  "frequency": "ONE_TIME",
  "methods": ["PIX"],
  "products": [{"external_id": "p1", "name": "Test", "quantity": 1, "price": 5000}],
  "coupon_code": "DESCONTO20"
}
```

When a valid coupon is provided:
- `amount` reflects the discounted total
- `original_amount` contains the original (pre-discount) total
- `coupon_code` stores the applied coupon code
- Installments are calculated based on the discounted amount
- The coupon's `redeems_count` is incremented

If the coupon is invalid, expired, or has no remaining uses, the request returns an error.

See [coupons.md](coupons.md) for details on creating and managing coupons.

## Installments

See [installments.md](installments.md) for details on installment calculation and management.

## Recurring Payments

See [recurring.md](recurring.md) for details on automatic recurring billing.
