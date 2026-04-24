# Billing

Cobranças são a principal entidade de pagamento. Elas representam uma cobrança a um cliente que pode ser paga via PIX ou cartão de crédito, opcionalmente parcelada com juros, e pode ser configurada para recorrência automática.

## Criar uma cobrança

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
  "external_id": "pedido-12345",
  "metadata": {
    "origem": "site",
    "canal": "organico"
  }
}
```

### Campos

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `frequency` | string | sim | `ONE_TIME` ou `MULTIPLE_PAYMENTS` |
| `methods` | string[] | sim | Métodos de pagamento: `PIX`, `CARD`, ou ambos |
| `products` | object[] | sim | Pelo menos um produto |
| `products[].external_id` | string | sim | Seu identificador do produto |
| `products[].name` | string | sim | Nome do produto |
| `products[].description` | string | não | Descrição do produto |
| `products[].quantity` | int | sim | Quantidade (deve ser >= 1) |
| `products[].price` | int | sim | Preço em centavos (deve ser >= 100) |
| `return_url` | string | não | URL para a ação de "voltar" após uma recusa |
| `completion_url` | string | não | URL de redirecionamento após aprovação |
| `customer_id` | string | não | ID de um cliente existente |
| `customer` | object | não | Dados do cliente (cria um novo cliente) |
| `installments` | int | não | Número de parcelas (1-12, padrão 1) |
| `interest_rate` | float | não | Taxa de juros mensal % (padrão: variável de ambiente `MOCKPAY_INTEREST_RATE`) |
| `coupon_code` | string | não | Código do cupom para aplicar desconto no valor total |
| `external_id` | string | não | Seu próprio identificador para esta cobrança |
| `metadata` | object | não | Pares chave-valor para dados adicionais |

O `amount` total da cobrança é a soma de todos os produtos (`price * quantity`).

### Resposta

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
    "external_id": "pedido-12345",
    "metadata": {
      "origem": "site",
      "canal": "organico"
    },
    "created_at": "2026-04-21T12:00:00.000",
    "updated_at": "2026-04-21T12:00:00.000"
  },
  "error": null
}
```

## Obter uma cobrança

```
GET /v1/billing/:id
```

## Listar cobranças

```
GET /v1/billing/list
```

Retorna todas as cobranças ordenadas por data de criação (mais recentes primeiro).

## Ciclo de status

```
PENDING → APPROVED
PENDING → DENIED
PENDING → EXPIRED (apenas recorrentes)
PENDING → CANCELLED
APPROVED → CANCELLED (apenas recorrentes)
```

- **APPROVED** - O pagamento foi aprovado via checkout ou dashboard.
- **DENIED** - O pagamento foi recusado via checkout ou dashboard.
- **CANCELLED** - A cobrança foi cancelada via API. Para cobranças recorrentes, interrompe cobranças futuras.
- **As transições de status são definitivas** - uma cobrança não pode ter seu status alterado após ser aprovada ou recusada (exceto cobranças recorrentes que podem transitar para CANCELLED).

## Cancelar uma cobrança

```
POST /v1/billing/:id/cancel
```

### Comportamento

- Cobranças **ONE_TIME**: apenas cobranças `PENDING` podem ser canceladas
- Cobranças **MULTIPLE_PAYMENTS** (recorrentes): cobranças `PENDING` e `APPROVED` podem ser canceladas
- Ao cancelar uma cobrança recorrente, o `next_billing` é limpo, interrompendo o ciclo de recorrência
- Todas as parcelas são marcadas como `CANCELLED`
- Um evento de webhook `billing.cancelled` é disparado

### Resposta

```json
{
  "data": {
    "id": "abc123...",
    "status": "CANCELLED",
    "frequency": "MULTIPLE_PAYMENTS",
    "next_billing": null,
    ...
  }
}
```

## Fluxo de aprovação

1. A cobrança é criada com o status `PENDING`
2. Uma página de checkout está disponível em `/checkout/<id>` com os botões Approve/Deny
3. Na aprovação: o status muda para `APPROVED`, todas as parcelas são marcadas como aprovadas, o webhook da `completion_url` é disparado
4. Na recusa: o status muda para `DENIED`, todas as parcelas são marcadas como recusadas, a `return_url` é usada para redirecionamento

## Cupom / Desconto

Um cupom pode ser aplicado na criação do billing passando `coupon_code`:

```json
{
  "frequency": "ONE_TIME",
  "methods": ["PIX"],
  "products": [{"external_id": "p1", "name": "Test", "quantity": 1, "price": 5000}],
  "coupon_code": "DESCONTO20"
}
```

Quando um cupom válido é fornecido:
- `amount` reflete o total com desconto
- `original_amount` contém o total original (sem desconto)
- `coupon_code` armazena o código do cupom aplicado
- As parcelas são calculadas com base no valor com desconto
- O `redeems_count` do cupom é incrementado

Se o cupom for inválido, expirado ou sem usos restantes, a requisição retorna um erro.

Veja [coupons.md](coupons.md) para detalhes sobre criação e gerenciamento de cupons.

## Parcelas

Consulte [installments.md](installments.md) para detalhes sobre cálculo e gerenciamento de parcelas.

## Pagamentos recorrentes

Consulte [recurring.md](recurring.md) para detalhes sobre cobranças recorrentes automáticas.
