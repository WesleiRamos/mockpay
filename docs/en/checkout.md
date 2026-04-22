# Checkout & Dashboard

MockPay includes a visual checkout page for each payment and a dashboard to manage all payments from the browser. These are HTML pages served without authentication — designed for development/testing workflows.

## Checkout Page

```
GET /checkout/{id}
```

Renders a two-column card layout:

### Left Panel — Order Summary
- MockPay branding
- Total amount (large display)
- Payment ID
- Status pill (PENDING / APPROVED / DENIED / EXPIRED)
- Payment method(s)
- Product list with quantities and prices
- "Simulated payment · testing only" disclaimer

### Right Panel — Payment Action

**When PENDING:**
- Method tabs (if both PIX and CARD are available)
- **PIX panel**: QR code image for scanning
- **CARD panel**: Card number, name, expiry, CVV fields, installment selector with dynamic calculation
- **Approve** button (green) and **Deny** button (red)

**When resolved (APPROVED/DENIED/EXPIRED):**
- Status icon and message
- No action buttons

### Mobile Responsive

On screens narrower than 680px, the layout switches from two columns to single column.

## Approve Payment

```
GET /checkout/{id}/approve
```

Approves the payment and redirects:
- For billings: redirects to `completion_url` (if set), otherwise back to checkout
- For PIX charges: redirects back to the checkout page

## Deny Payment

```
GET /checkout/{id}/deny
```

Denies the payment and redirects:
- For billings: redirects to `return_url` (if set), otherwise back to checkout
- For PIX charges: redirects back to the checkout page

## QR Code Integration

When a payment is PENDING, the checkout page generates a real QR code (via `go-qrcode`) containing the approve URL:

```
http://<hostname>:<port>/checkout/<id>/approve
```

Scanning this QR code with a phone opens the approve endpoint directly, confirming the payment with a single action.

## Dashboard

```
GET /
```

Interactive dashboard showing all payments across billings and PIX charges.

### Features

- **Stat cards** — Total, Pending, Approved, Denied/Expired counts
- **Clickable filters** — Click a stat card to filter the table by status
- **Transactions table** — ID, Amount, Status, Method, Type (billing/pix), Created date, View link
- **Status pills** — Color-coded status badges for each transaction
- **Type chips** — Distinguish between `billing` and `pix` payment types
- **Auto-refresh** — Page reloads every 5 seconds with countdown indicator
- **Fade-up animations** — Staggered row animations for visual polish

### Design

The dashboard follows the design system:
- Inter font family
- Light theme with `#f7f8f5` background
- Green accent (`#9fe870`)
- Rounded cards and pill-shaped buttons
- Sticky navigation bar with centered logo

## Health Check

```
GET /health
```

Returns JSON:

```json
{
  "status": "ok",
  "timestamp": "2026-04-21T12:00:00.000"
}
```

## Statistics

```
GET /v1/stats
```

Requires authentication. Returns aggregate counts:

```json
{
  "data": {
    "billings_total": 5,
    "billings_pending": 2,
    "billings_approved": 2,
    "billings_denied": 1,
    "pix_total": 3,
    "pix_pending": 1,
    "pix_approved": 1,
    "pix_expired": 1,
    "customers_total": 4,
    "coupons_total": 2
  },
  "error": null
}
```
