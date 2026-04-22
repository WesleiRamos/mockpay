# Coupons

Coupons provide discount functionality that can be applied to billing amounts. They support percentage-based or fixed-amount discounts with configurable redemption limits.

## Create a Coupon

```
POST /v1/coupon/create
```

```json
{
  "code": "DESCONTO20",
  "notes": "20% discount coupon",
  "max_redeems": 10,
  "discount_kind": "PERCENTAGE",
  "discount": 20
}
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `code` | string | yes | Unique coupon code |
| `notes` | string | no | Description or notes |
| `max_redeems` | int | yes | Maximum number of redemptions (-1 for unlimited) |
| `discount_kind` | string | yes | `PERCENTAGE` or `FIXED` |
| `discount` | int | yes | Discount value (percentage number or cents for fixed) |

### Response

```json
{
  "data": {
    "id": "a1b2c3d4...",
    "code": "DESCONTO20",
    "notes": "20% discount coupon",
    "max_redeems": 10,
    "redeems_count": 0,
    "discount_kind": "PERCENTAGE",
    "discount": 20,
    "dev_mode": true,
    "status": "ACTIVE",
    "created_at": "2026-04-21T12:00:00.000",
    "updated_at": "2026-04-21T12:00:00.000"
  },
  "error": null
}
```

## List Coupons

```
GET /v1/coupon/list
```

Returns all coupons ordered by creation date (newest first).

## Discount Types

### Percentage

```json
{
  "discount_kind": "PERCENTAGE",
  "discount": 20
}
```

Applies a percentage discount. For a R$ 100,00 (10000 cents) billing:

```
10000 - (10000 * 20 / 100) = 8000 cents (R$ 80,00)
```

### Fixed

```json
{
  "discount_kind": "FIXED",
  "discount": 500
}
```

Subtracts a fixed amount in cents. For a R$ 50,00 (5000 cents) billing:

```
5000 - 500 = 4500 cents (R$ 45,00)
```

## Redemption Limits

- `max_redeems` controls how many times a coupon can be used
- `-1` means unlimited redemptions
- Each call to `ApplyDiscount` increments `redeems_count`
- Once `redeems_count >= max_redeems`, the coupon can no longer be used (unless `max_redeems` is -1)

## Coupon Status

| Status | Description |
|--------|-------------|
| `ACTIVE` | Coupon is available for use |
| `INACTIVE` | Coupon is disabled |

New coupons are created with `ACTIVE` status.

## Apply Discount

The `ApplyDiscount` function in `CouponService` validates and applies the coupon:

1. Checks the coupon exists
2. Checks it has `ACTIVE` status
3. Checks redemption limit (`redeems_count < max_redeems` or `max_redeems == -1`)
4. Calculates the discounted amount
5. Caps the discount at the full amount (result can't go below 0)
6. Increments `redeems_count`
7. Returns the final discounted amount in cents

## Using Coupons in Billing

Coupons are applied when creating a billing via the `coupon_code` field:

```bash
curl -X POST http://localhost:8080/v1/billing/create \
  -H "Authorization: Bearer mock_key" \
  -H "Content-Type: application/json" \
  -d '{
    "frequency": "ONE_TIME",
    "methods": ["PIX"],
    "products": [{"external_id": "p1", "name": "Test Product", "quantity": 1, "price": 5000}],
    "coupon_code": "DESCONTO20"
  }'
```

The billing response will include:
- `amount` — the discounted total
- `original_amount` — the original total before discount
- `coupon_code` — the applied coupon code

See [billing.md](billing.md) for full billing creation details.
