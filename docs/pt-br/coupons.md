# Coupons

Coupons fornecem funcionalidade de desconto que pode ser aplicada a valores de cobrança. Eles suportam descontos por porcentagem ou valor fixo com limites de resgate configuráveis.

## Criar um Coupon

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

### Campos

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `code` | string | sim | Código único do coupon |
| `notes` | string | não | Descrição ou observações |
| `max_redeems` | int | sim | Número máximo de resgates (-1 para ilimitado) |
| `discount_kind` | string | sim | `PERCENTAGE` ou `FIXED` |
| `discount` | int | sim | Valor do desconto (número da porcentagem ou centavos para valor fixo) |

### Resposta

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

## Listar Coupons

```
GET /v1/coupon/list
```

Retorna todos os coupons ordenados por data de criação (mais recentes primeiro).

## Tipos de Desconto

### Porcentagem

```json
{
  "discount_kind": "PERCENTAGE",
  "discount": 20
}
```

Aplica um desconto por porcentagem. Para uma cobrança de R$ 100,00 (10000 centavos):

```
10000 - (10000 * 20 / 100) = 8000 centavos (R$ 80,00)
```

### Fixo

```json
{
  "discount_kind": "FIXED",
  "discount": 500
}
```

Subtrai um valor fixo em centavos. Para uma cobrança de R$ 50,00 (5000 centavos):

```
5000 - 500 = 4500 centavos (R$ 45,00)
```

## Limites de Resgate

- `max_redeems` controla quantas vezes um coupon pode ser usado
- `-1` significa resgates ilimitados
- Cada chamada a `ApplyDiscount` incrementa `redeems_count`
- Quando `redeems_count >= max_redeems`, o coupon não pode mais ser utilizado (a menos que `max_redeems` seja -1)

## Status do Coupon

| Status | Descrição |
|--------|-----------|
| `ACTIVE` | Coupon está disponível para uso |
| `INACTIVE` | Coupon está desativado |

Novos coupons são criados com o status `ACTIVE`.

## Aplicar Desconto

A função `ApplyDiscount` em `CouponService` valida e aplica o coupon:

1. Verifica se o coupon existe
2. Verifica se possui o status `ACTIVE`
3. Verifica o limite de resgate (`redeems_count < max_redeems` ou `max_redeems == -1`)
4. Calcula o valor com desconto
5. Limita o desconto ao valor total (o resultado não pode ser menor que 0)
6. Incrementa `redeems_count`
7. Retorna o valor final com desconto em centavos

## Usando Cupons no Billing

Os cupons são aplicados ao criar um billing através do campo `coupon_code`:

```bash
curl -X POST http://localhost:8080/v1/billing/create \
  -H "Authorization: Bearer mock_key" \
  -H "Content-Type: application/json" \
  -d '{
    "frequency": "ONE_TIME",
    "methods": ["PIX"],
    "products": [{"external_id": "p1", "name": "Produto Teste", "quantity": 1, "price": 5000}],
    "coupon_code": "DESCONTO20"
  }'
```

A resposta do billing incluirá:
- `amount` — o total com desconto
- `original_amount` — o total original antes do desconto
- `coupon_code` — o código do cupom aplicado

Veja [billing.md](billing.md) para detalhes completos da criação de billings.
