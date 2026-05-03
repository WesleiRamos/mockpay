# Webhooks

O MockPay pode enviar notificações reais de webhook via HTTP quando eventos de pagamento ocorrem. Os webhooks estão desabilitados por padrão e devem ser configurados através de variáveis de ambiente.

## Configuração

```bash
MOCKPAY_WEBHOOK_URL=https://your-server.com/webhooks
MOCKPAY_WEBHOOK_SECRET=your-secret-key
```

Ambas as variáveis são opcionais. Se `MOCKPAY_WEBHOOK_URL` estiver vazia, os webhooks ficam desabilitados.

## Eventos

| Evento | Gatilho |
|--------|---------|
| `billing.approved` | Pagamento de cobrança aprovado |
| `billing.denied` | Pagamento de cobrança negado |
| `billing.cancelled` | Cobrança cancelada via API |
| `billing.created` | Cobrança recorrente criada automaticamente |
| `pix.approved` | Pagamento PIX aprovado |
| `pix.expired` | Cobrança PIX expirada |
| `pix.payout.liquidated` | PIX payout (saque) liquidado |

## Payload

Cada requisição de webhook é um `POST` com `Content-Type: application/json`:

```json
{
  "id": "evt_xxxx",
  "type": "billing.approved",
  "payload": {
    "id": "abc123...",
    "amount": 5000,
    "status": "APPROVED",
    ...
  },
  "created_at": "2026-04-21T12:00:00.000"
}
```

O `payload` contém a entidade completa (cobrança ou charge PIX) no momento do evento.

## Verificação de Assinatura

Cada requisição de webhook inclui um header `X-MockPay-Signature`:

```
X-MockPay-Signature: sha256=<hex_digest>
```

A assinatura é calculada como **HMAC-SHA256** do corpo bruto da requisição usando o segredo do webhook.

### Verificando na Sua Aplicação

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "strings"
)

func verifySignature(secret string, body []byte, header string) bool {
    if !strings.HasPrefix(header, "sha256=") {
        return false
    }
    sig, _ := hex.DecodeString(header[7:])

    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(body)
    expected := mac.Sum(nil)

    return hmac.Equal(sig, expected)
}
```

### Verificando com curl

```bash
# Test your webhook receiver locally
# Using webhook.site or similar services for testing

MOCKPAY_WEBHOOK_URL=https://webhook.site/your-unique-url
MOCKPAY_WEBHOOK_SECRET=my-secret
```

## Entrega e Retentativa

Os webhooks são enviados de forma assíncrona através de um channel com buffer (capacidade 100):

1. **Envio inicial** - HTTP POST com timeout de 5 segundos
2. **Em caso de falha** (resposta não-2xx ou erro de rede):
   - Registra a tentativa de entrega no banco de dados
   - Enfileira para retentativa
3. **Retentativas** - Até 3 tentativas no total com backoff linear:
   - Tentativa 1: imediata
   - Tentativa 2: após 5 segundos
   - Tentativa 3: após 10 segundos

### Registros de Entrega

Cada tentativa de entrega é armazenada na tabela `webhook_deliveries` com:
- `id` - ID único da entrega
- `event_id` - o evento que está sendo entregue
- `url` - a URL de destino
- `attempt` - número da tentativa (1-3)
- `status_code` - código de resposta HTTP (0 em caso de erro de rede)
- `success` - se a entrega foi bem-sucedida (resposta 2xx)

## Arquitetura

O sistema de webhooks utiliza duas goroutines em segundo plano:

1. **Worker** - lê do channel de eventos e envia as requisições HTTP
2. **Retry Worker** - lê do channel de retentativas, aplica o delay de backoff e reenvia

```
Dispatch() → events channel → worker → HTTP POST
                                         ↓ failure
                                    retries channel → retryWorker → HTTP POST
```

Este design garante que a entrega de webhooks não bloqueie o tratamento principal das requisições.
