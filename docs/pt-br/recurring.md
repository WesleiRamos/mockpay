# Pagamentos Recorrentes

Cobranças com `frequency: "MULTIPLE_PAYMENTS"` geram automaticamente novas cobranças de forma recorrente.

## Como Funciona

1. Crie uma cobrança com `"frequency": "MULTIPLE_PAYMENTS"` e os produtos/métodos desejados
2. A cobrança recebe um campo `next_billing` definido para **30 dias a partir da criação**
3. Uma tarefa em segundo plano verifica a cada **30 segundos** se há cobranças recorrentes cuja data de `next_billing` já passou
4. Quando vencida, uma **nova cobrança** é criada automaticamente com:
   - Mesmos produtos, métodos e cliente
   - Status `PENDING`
   - Um novo ID único
   - Seu próprio `next_billing` definido para 30 dias à frente
5. O `next_billing` da cobrança original é atualizado para o próximo ciclo
6. Um evento de webhook `billing.created` é disparado
7. O ciclo se repete indefinidamente

## Criando uma Cobrança Recorrente

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

### Resposta

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

O campo `next_billing` indica quando a próxima cobrança automática será gerada.

## Tipos de Frequência

| Frequência | Comportamento |
|-----------|----------|
| `ONE_TIME` | Pagamento único, sem recorrência. `next_billing` é nulo. |
| `MULTIPLE_PAYMENTS` | Gera uma nova cobrança a cada 30 dias. `next_billing` é definido. |

## Tarefa em Segundo Plano

A verificação recorrente é executada em uma goroutine iniciada por `StartBackgroundJobs()`:

```go
// Runs every 30 seconds
ticker := time.NewTicker(30 * time.Second)
for range ticker.C {
    ps.CheckExpired()      // expires PIX charges
    bs.CheckRecurring()    // generates recurring billings
}
```

### Lógica do CheckRecurring

1. Lista todas as cobranças com status `PENDING` e frequência `MULTIPLE_PAYMENTS`
2. Para cada uma, analisa o timestamp de `next_billing`
3. Se `next_billing` está no passado:
   - Cria uma nova cobrança com os mesmos produtos/métodos/cliente
   - Atualiza o `next_billing` da cobrança original para +30 dias
   - Dispara o webhook `billing.created`

## Eventos de Webhook

| Evento | Gatilho |
|-------|---------|
| `billing.created` | Uma nova cobrança recorrente é gerada automaticamente |

O payload do webhook contém o objeto completo da nova cobrança.

## Interrompendo a Recorrência

Atualmente, a recorrência continua até que:
- A cobrança original seja aprovada ou negada (o status muda de PENDING)
- O servidor seja interrompido

Não existe um endpoint explícito de cancelamento para cobranças recorrentes.
