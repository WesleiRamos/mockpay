# Clientes

Clientes representam as pessoas que realizam pagamentos. Eles podem ser criados explicitamente ou implicitamente durante a criacao de cobrancas/PIX.

## Criar um Cliente

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

### Campos

| Campo | Tipo | Obrigatorio | Descricao |
|-------|------|-------------|-----------|
| `name` | string | nao | Nome do cliente |
| `cellphone` | string | nao | Numero de telefone |
| `email` | string | sim | Endereco de email (usado para deduplicacao) |
| `tax_id` | string | nao | CPF/CNPJ |

### Resposta

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

### Deduplicacao por Email

Se um cliente com o mesmo `email` ja existir, o cliente existente sera retornado em vez de criar um duplicado. Isso torna o endpoint de criacao **idempotente** por email.

## Listar Clientes

```
GET /v1/customer/list
```

Retorna todos os clientes ordenados por data de criacao (mais recentes primeiro).

## Criacao Implicita de Clientes

Clientes sao criados automaticamente quando:

1. Uma cobranca e criada com um objeto `customer` (e sem `customer_id`)
2. Uma cobranca PIX e criada com um objeto `customer`

Em ambos os casos, a mesma deduplicacao por email se aplica — se um cliente com esse email ja existir, ele sera reaproveitado.

## Referencias de Cliente

Quando um cliente e vinculado a uma cobranca ou cobranca PIX, uma referencia leve `CustomerRef` e armazenada:

```json
{
  "id": "a1b2c3d4...",
  "metadata": {
    "name": "Joao Silva",
    "email": "joao@email.com"
  }
}
```

## Armazenamento de Dados

Os dados do cliente sao armazenados como um mapa JSON `metadata` (`map[string]string`). Esse esquema flexivel permite armazenar quaisquer pares de chave-valor sem necessidade de migracoes de esquema. As chaves padrao preenchidas a partir de `CustomerInput` sao:

- `name`
- `email`
- `cellphone`
- `tax_id`
