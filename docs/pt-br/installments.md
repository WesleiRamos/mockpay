# Parcelamento

Cobranças podem ser divididas em múltiplas parcelas com juros opcionais. Isso é usado principalmente com pagamentos via cartão de crédito.

## Criando com Parcelamento

Defina `installments` (> 1) e, opcionalmente, `interest_rate` na requisição de criação da cobrança:

```json
{
  "frequency": "ONE_TIME",
  "methods": ["CARD"],
  "products": [{"external_id": "p1", "name": "Test", "quantity": 1, "price": 10000}],
  "installments": 3,
  "interest_rate": 5.0
}
```

### Campos

| Campo | Tipo | Padrão | Descrição |
|-------|------|--------|-----------|
| `installments` | int | 1 | Número de parcelas (1-12). 1 significa sem parcelamento (pagamento à vista). |
| `interest_rate` | float | variável de ambiente `MOCKPAY_INTEREST_RATE` (padrão 0) | Taxa de juros mensal em %. |

## Cálculo de Juros

O MockPay utiliza **juros simples**:

```
total = amount * (1 + rate/100 * installments)
```

### Exemplo

R$ 100,00 (10000 centavos) em 3x com juros mensais de 5%:

- Total: `10000 * (1 + 5/100 * 3)` = `10000 * 1.15` = **11500 centavos (R$ 115,00)**
- Por parcela: `11500 / 3` = **3833 centavos** (R$ 38,33)
- O resto é adicionado à última parcela: `11500 - 3833 * 2` = **3834 centavos** (R$ 38,34)

Resultado:
| # | Valor | Status | Vencimento |
|---|-------|--------|------------|
| 1 | R$ 38,33 | PENDING | +30 dias |
| 2 | R$ 38,33 | PENDING | +60 dias |
| 3 | R$ 38,34 | PENDING | +90 dias |

### Datas de Vencimento

As datas de vencimento das parcelas são espaçadas em intervalos de 30 dias, começando a partir da data de criação da cobrança.

### Tratamento do Resto

A última parcela absorve qualquer resto da divisão inteira, garantindo que a soma de todas as parcelas seja exatamente igual ao total com juros.

## Listando Parcelas

```
GET /v1/billing/{id}/installments
```

Retorna a lista de parcelas de uma cobrança:

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

## Comportamento na Aprovação

Quando uma cobrança é aprovada, **todas as parcelas são aprovadas juntas**:

- Todos os status das parcelas mudam para `APPROVED`
- `paid_at` é definido como o timestamp atual para cada parcela

Quando uma cobrança é negada, todas as parcelas são marcadas como negadas.

## Taxa de Juros Padrão

Se `interest_rate` não for informado na requisição, o valor da variável de ambiente `MOCKPAY_INTEREST_RATE` será utilizado (padrão: 0).
