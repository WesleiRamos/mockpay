# Installments

Billings can be split into multiple installments with optional interest. This is primarily used with credit card payments.

## Creating with Installments

Set `installments` (> 1) and optionally `interest_rate` in the billing creation request:

```json
{
  "frequency": "ONE_TIME",
  "methods": ["CARD"],
  "products": [{"external_id": "p1", "name": "Test", "quantity": 1, "price": 10000}],
  "installments": 3,
  "interest_rate": 5.0
}
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `installments` | int | 1 | Number of installments (1-12). 1 means no installments (lump sum). |
| `interest_rate` | float | `MOCKPAY_INTEREST_RATE` env var (default 0) | Monthly interest rate in %. |

## Interest Calculation

MockPay uses **simple interest**:

```
total = amount * (1 + rate/100 * installments)
```

### Example

R$ 100,00 (10000 cents) in 3x with 5% monthly interest:

- Total: `10000 * (1 + 5/100 * 3)` = `10000 * 1.15` = **11500 cents (R$ 115,00)**
- Per installment: `11500 / 3` = **3833 cents** (R$ 38,33)
- Remainder goes to last installment: `11500 - 3833 * 2` = **3834 cents** (R$ 38,34)

Result:
| # | Amount | Status | Due Date |
|---|--------|--------|----------|
| 1 | R$ 38,33 | PENDING | +30 days |
| 2 | R$ 38,33 | PENDING | +60 days |
| 3 | R$ 38,34 | PENDING | +90 days |

### Due Dates

Installment due dates are spaced 30 days apart starting from the billing creation date.

### Remainder Handling

The last installment absorbs any remainder from integer division, ensuring the sum of all installments equals exactly the total with interest.

## Listing Installments

```
GET /v1/billing/{id}/installments
```

Returns the installment list for a billing:

```json
{
  "data": [
    {
      "number": 1,
      "amount": 3833,
      "status": "PENDING",
      "due_date": "2026-05-21T12:00:00.000",
      "paid_at": null
    },
    {
      "number": 2,
      "amount": 3833,
      "status": "PENDING",
      "due_date": "2026-06-20T12:00:00.000",
      "paid_at": null
    },
    {
      "number": 3,
      "amount": 3834,
      "status": "PENDING",
      "due_date": "2026-07-20T12:00:00.000",
      "paid_at": null
    }
  ],
  "error": null
}
```

## Approval Behavior

When a billing is approved, **all installments are approved together**:

- All installment statuses change to `APPROVED`
- `paid_at` is set to the current timestamp for each installment

When a billing is denied, all installments are marked as denied.

## Default Interest Rate

If `interest_rate` is not provided in the request, the value from the `MOCKPAY_INTEREST_RATE` environment variable is used (default: 0).
