# Storage

MockPay uses SQLite for persistence via the pure-Go driver `modernc.org/sqlite` (no CGO required). Data survives server restarts.

## Configuration

```bash
MOCKPAY_DB_PATH=mockpay.db   # default
```

Set `MOCKPAY_DB_PATH` to customize the database file location. Use `:memory:` for in-memory mode (useful for testing).

## Database Schema

### billings

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT PK | 32-char hex ID |
| `url` | TEXT | Checkout URL |
| `amount` | INTEGER | Total in cents |
| `status` | TEXT | PENDING, APPROVED, DENIED, CANCELLED, EXPIRED |
| `dev_mode` | INTEGER | Boolean (0/1) |
| `methods` | TEXT | JSON array of payment methods |
| `products` | TEXT | JSON array of product objects |
| `frequency` | TEXT | ONE_TIME or MULTIPLE_PAYMENTS |
| `next_billing` | TEXT | Next recurring billing date (nullable) |
| `customer` | TEXT | JSON customer reference (nullable) |
| `return_url` | TEXT | Redirect URL on denial |
| `completion_url` | TEXT | Redirect URL on approval |
| `installments` | INTEGER | Number of installments |
| `interest_rate` | REAL | Monthly interest rate % |
| `installment_list` | TEXT | JSON array of installment objects |
| `created_at` | TEXT | Creation timestamp |
| `updated_at` | TEXT | Last update timestamp |

### customers

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT PK | 32-char hex ID |
| `metadata` | TEXT | JSON object with customer fields |
| `created_at` | TEXT | Creation timestamp |
| `updated_at` | TEXT | Last update timestamp |

### coupons

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT PK | 32-char hex ID |
| `code` | TEXT | Unique coupon code |
| `notes` | TEXT | Description |
| `max_redeems` | INTEGER | Maximum redemptions |
| `redeems_count` | INTEGER | Current redemption count |
| `discount_kind` | TEXT | PERCENTAGE or FIXED |
| `discount` | INTEGER | Discount value |
| `dev_mode` | INTEGER | Boolean (0/1) |
| `status` | TEXT | ACTIVE or INACTIVE |
| `created_at` | TEXT | Creation timestamp |
| `updated_at` | TEXT | Last update timestamp |

### pix_charges

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT PK | 32-char hex ID |
| `amount` | INTEGER | Amount in cents |
| `status` | TEXT | PENDING, APPROVED, EXPIRED |
| `dev_mode` | INTEGER | Boolean (0/1) |
| `br_code` | TEXT | QR code URL |
| `br_code_base64` | TEXT | Base64 data URI of QR PNG |
| `platform_fee` | INTEGER | Platform fee in cents |
| `expires_at` | TEXT | Expiration timestamp |
| `customer` | TEXT | JSON customer reference (nullable) |
| `created_at` | TEXT | Creation timestamp |
| `updated_at` | TEXT | Last update timestamp |

### webhook_deliveries

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT PK | 32-char hex ID |
| `event_id` | TEXT | Associated event ID |
| `url` | TEXT | Target webhook URL |
| `attempt` | INTEGER | Attempt number (1-3) |
| `status_code` | INTEGER | HTTP response code |
| `success` | INTEGER | Boolean (0/1) |
| `created_at` | TEXT | Delivery timestamp |

## JSON Columns

Complex fields are serialized as JSON text:

- `methods` → `["PIX","CARD"]`
- `products` → `[{"external_id":"p1","name":"Test","quantity":1,"price":5000}]`
- `installment_list` → `[{"number":1,"amount":1750,"status":"PENDING","due_date":"..."}]`
- `customer` → `{"id":"abc","metadata":{"email":"test@test.com"}}`
- `metadata` → `{"name":"Joao","email":"joao@test.com"}`

## Concurrency

The store uses `sync.RWMutex` for thread-safe access:
- Write operations (`Create*`, `Update*`) acquire an exclusive lock
- Read operations (`Get*`, `List*`) acquire a shared lock
- SQLite connection is limited to 1 max open connection for write safety

## Migrations

Schema migrations run automatically on startup via `CREATE TABLE IF NOT EXISTS`. No migration tool is needed — the database is created from scratch if it doesn't exist.

## Data Cleanup

To reset all data, simply delete the database file and restart:

```bash
rm mockpay.db
go run main.go
```
