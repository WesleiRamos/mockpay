# PIX Payout

PIX Payouts representam solicitações de saque (PIX Send). Cada payout é criado com status `PROCESSING` e é liquidado automaticamente após 30 segundos, simulando o fluxo de transferência PIX. Um webhook é disparado quando a liquidação é concluída.

## Criar um PIX Payout

```
POST /v1/pix/payouts
```

```json
{
  "amount": 5000,
  "pix_key_type": "cpf",
  "pix_key": "12345678901",
  "external_id": "019b38f7-2b67-7e0a-a8f3-5c4f72b61f1d",
  "idempotency_key": "pix-withdrawal:payout:019b38f7-2b67-7e0a-a8f3-5c4f72b61f1d",
  "metadata": {
    "source": "wallet"
  }
}
```

### Campos

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `amount` | int | sim | Valor em centavos (deve ser > 0) |
| `pix_key_type` | string | sim | Tipo da chave PIX: `cpf`, `phone`, `email` ou `random` |
| `pix_key` | string | sim | Chave PIX do recebedor |
| `external_id` | string | não | Seu próprio identificador para este payout |
| `idempotency_key` | string | sim | Chave de idempotência gerada pelo chamador |
| `metadata` | object | não | Pares chave-valor para dados adicionais |

### Resposta

```json
{
  "data": {
    "id": "payout_01HXZ6R0V7V8D8QN7M5W6B8J9K",
    "amount": 5000,
    "status": "PROCESSING",
    "external_id": "019b38f7-2b67-7e0a-a8f3-5c4f72b61f1d",
    "end_to_end_id": "",
    "metadata": {
      "source": "wallet"
    },
    "created_at": "2026-05-03T12:00:00.000",
    "updated_at": "2026-05-03T12:00:00.000"
  },
  "error": null
}
```

## Verificar Status do Payout

```
GET /v1/pix/payouts/:id/check
```

Retorna o payout atual com seu status.

### Resposta em processamento

```json
{
  "data": {
    "id": "payout_01HXZ6R0V7V8D8QN7M5W6B8J9K",
    "amount": 5000,
    "status": "PROCESSING",
    "external_id": "019b38f7-2b67-7e0a-a8f3-5c4f72b61f1d",
    "end_to_end_id": "",
    "metadata": {
      "source": "wallet"
    },
    "created_at": "2026-05-03T12:00:00.000",
    "updated_at": "2026-05-03T12:00:00.000"
  },
  "error": null
}
```

### Resposta liquidado

```json
{
  "data": {
    "id": "payout_01HXZ6R0V7V8D8QN7M5W6B8J9K",
    "amount": 5000,
    "status": "LIQUIDATED",
    "external_id": "019b38f7-2b67-7e0a-a8f3-5c4f72b61f1d",
    "end_to_end_id": "E1823612020260503120000000000001",
    "metadata": {
      "source": "wallet"
    },
    "created_at": "2026-05-03T12:00:00.000",
    "updated_at": "2026-05-03T12:01:00.000"
  },
  "error": null
}
```

## Ciclo de Status

```
PROCESSING → LIQUIDATED   (automático, após 30 segundos)
```

As carteiras devem interpretar os seguintes status:
- `PROCESSING` — saque enviado, aguardando liquidação
- `LIQUIDATED` ou `PAID` — saque liquidado com sucesso
- `FAILED`, `REFUSED`, `REJECTED` ou `CANCELLED` — saque falhou, a carteira libera a reserva

## Liquidação Automática

Uma tarefa em segundo plano executa a cada 30 segundos para verificar payouts pendentes:

1. Lista todos os payouts com status `PROCESSING`
2. Verifica se 30 segundos se passaram desde a criação
3. Se elegível, altera o status para `LIQUIDATED`
4. Gera um `end_to_end_id` (formato: `E20260502...`)
5. Dispara o webhook `pix.payout.liquidated`

## Idempotência

- `POST /v1/pix/payouts` com a mesma `idempotency_key` retorna o mesmo payout
- `GET /v1/pix/payouts/{payout_id}/check` retorna o estado atual sem criar outro payout
- Webhooks repetidos com o mesmo `id` podem ser reenviados; a carteira deduplica o evento
- O `amount` retornado no payout precisa ser igual ao `amount` solicitado, senão a carteira rejeita a liquidação

## Webhook

Quando o payout é liquidado, o MockPay envia um `POST` para a URL configurada em `MOCKPAY_WEBHOOK_URL`:

```json
{
  "id": "evt_payout_01HXZ6R0V7V8D8QN7M5W6B8J9K",
  "type": "pix.payout.liquidated",
  "payload": {
    "id": "payout_01HXZ6R0V7V8D8QN7M5W6B8J9K",
    "amount": 5000,
    "status": "LIQUIDATED",
    "external_id": "019b38f7-2b67-7e0a-a8f3-5c4f72b61f1d",
    "end_to_end_id": "E1823612020260503120000000000001",
    "metadata": {
      "source": "wallet"
    },
    "created_at": "2026-05-03T12:00:00.000",
    "updated_at": "2026-05-03T12:01:00.000"
  },
  "created_at": "2026-05-03T12:01:00.000"
}
```

O header `X-MockPay-Signature` é incluído com HMAC-SHA256 quando `MOCKPAY_WEBHOOK_SECRET` está configurado.

## Exemplos com cURL

```bash
# Criar um payout
curl -X POST "http://localhost:8080/v1/pix/payouts" \
  -H "Authorization: Bearer mock_key" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 5000,
    "pix_key_type": "cpf",
    "pix_key": "12345678901",
    "external_id": "019b38f7-2b67-7e0a-a8f3-5c4f72b61f1d",
    "idempotency_key": "pix-withdrawal:payout:019b38f7-2b67-7e0a-a8f3-5c4f72b61f1d",
    "metadata": {
      "source": "wallet"
    }
  }'

# Verificar status do payout
curl -X GET "http://localhost:8080/v1/pix/payouts/payout_01HXZ6R0V7V8D8QN7M5W6B8J9K/check" \
  -H "Authorization: Bearer mock_key"
```
