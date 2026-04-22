# Customers

Customers represent the people making payments. They can be created explicitly or implicitly during billing/PIX creation.

## Create a Customer

```
POST /v1/customer/create
```

```json
{
  "name": "Joao Silva",
  "cellphone": "11999999999",
  "email": "joao@email.com",
  "tax_id": "12345678901"
}
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | no | Customer name |
| `cellphone` | string | no | Phone number |
| `email` | string | yes | Email address (used for deduplication) |
| `tax_id` | string | no | CPF/CNPJ |

### Response

```json
{
  "data": {
    "id": "a1b2c3d4...",
    "metadata": {
      "name": "Joao Silva",
      "email": "joao@email.com",
      "cellphone": "11999999999",
      "tax_id": "12345678901"
    },
    "created_at": "2026-04-21T12:00:00.000",
    "updated_at": "2026-04-21T12:00:00.000"
  },
  "error": null
}
```

### Email Deduplication

If a customer with the same `email` already exists, the existing customer is returned instead of creating a duplicate. This makes the create endpoint **idempotent** by email.

## List Customers

```
GET /v1/customer/list
```

Returns all customers ordered by creation date (newest first).

## Implicit Customer Creation

Customers are automatically created when:

1. A billing is created with a `customer` object (and no `customer_id`)
2. A PIX charge is created with a `customer` object

In both cases, the same email deduplication applies — if a customer with that email exists, it's reused.

## Customer References

When a customer is attached to a billing or PIX charge, a lightweight `CustomerRef` is stored:

```json
{
  "id": "a1b2c3d4...",
  "metadata": {
    "name": "Joao Silva",
    "email": "joao@email.com"
  }
}
```

## Data Storage

Customer data is stored as a `metadata` JSON map (`map[string]string`). This flexible schema allows any key-value pairs to be stored without schema migrations. The standard keys populated from `CustomerInput` are:

- `name`
- `email`
- `cellphone`
- `tax_id`
