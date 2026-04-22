# Recurring Payments

Billings with `frequency: "MULTIPLE_PAYMENTS"` automatically generate new billings on a recurring basis.

## How It Works

1. Create a billing with `"frequency": "MULTIPLE_PAYMENTS"` and the desired products/methods
2. The billing receives a `next_billing` field set to **30 days from creation**
3. A background job checks every **30 seconds** for recurring billings whose `next_billing` date has passed
4. When due, a **new billing** is auto-created with:
   - Same products, methods, and customer
   - Status `PENDING`
   - A new unique ID
   - Its own `next_billing` set to 30 days out
5. The original billing's `next_billing` is updated to the next cycle
6. A `billing.created` webhook event is dispatched
7. The cycle repeats indefinitely

## Creating a Recurring Billing

```bash
curl -X POST http://localhost:8080/v1/billing/create \
  -H "Authorization: Bearer mock_key" \
  -H "Content-Type: application/json" \
  -d '{
    "frequency": "MULTIPLE_PAYMENTS",
    "methods": ["PIX"],
    "products": [{"external_id": "sub1", "name": "Monthly Subscription", "quantity": 1, "price": 9900}]
  }'
```

### Response

```json
{
  "data": {
    "id": "abc123...",
    "amount": 9900,
    "status": "PENDING",
    "frequency": "MULTIPLE_PAYMENTS",
    "next_billing": "2026-05-21T12:00:00.000",
    "products": [...],
    ...
  }
}
```

The `next_billing` field indicates when the next automatic billing will be generated.

## Frequency Types

| Frequency | Behavior |
|-----------|----------|
| `ONE_TIME` | Single payment, no recurrence. `next_billing` is null. |
| `MULTIPLE_PAYMENTS` | Generates a new billing every 30 days. `next_billing` is set. |

## Background Job

The recurring check runs in a goroutine started by `StartBackgroundJobs()`:

```go
// Runs every 30 seconds
ticker := time.NewTicker(30 * time.Second)
for range ticker.C {
    ps.CheckExpired()      // expires PIX charges
    bs.CheckRecurring()    // generates recurring billings
}
```

### CheckRecurring Logic

1. Lists all billings with status `PENDING` and frequency `MULTIPLE_PAYMENTS`
2. For each, parses `next_billing` timestamp
3. If `next_billing` is in the past:
   - Creates a new billing with the same products/methods/customer
   - Updates the original billing's `next_billing` to +30 days
   - Dispatches `billing.created` webhook

## Webhook Events

| Event | Trigger |
|-------|---------|
| `billing.created` | A new recurring billing is auto-generated |

The webhook payload contains the full new billing object.

## Stopping Recurrence

Currently, recurrence continues until:
- The original billing is approved or denied (status changes from PENDING)
- The server is stopped

There is no explicit cancel endpoint for recurring billings.
