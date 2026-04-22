# Primeiros Passos

MockPay e um simulador de gateway de pagamento para desenvolvimento e testes. Ele simula fluxos de pagamento brasileiros (PIX, cartao de credito, parcelamentos, cupons) com um dashboard visual e interface de checkout.

## Requisitos

- Go 1.26+

## Instalacao e Execucao

```bash
git clone git@github.com:WesleiRamos/mockpay.git && cd mockpay
go mod download
go run main.go
```

O servidor inicia em `http://localhost:8080`.

## Configuracao

Crie um arquivo `.env` a partir do exemplo:

```bash
cp .env.example .env
```

| Variavel | Padrao | Descricao |
|----------|--------|-----------|
| `MOCKPAY_PORT` | `8080` | Porta HTTP de escuta |
| `MOCKPAY_API_KEY` | `mock_key` | Token Bearer para autenticacao na API |
| `MOCKPAY_BASE_URL` | `http://localhost:<port>` | URL base utilizada para construir links de checkout |
| `MOCKPAY_PUBLIC_URL` | `MOCKPAY_BASE_URL` | URL publica utilizada em QR codes (defina como o IP da sua rede local para testes em dispositivos moveis, utiliza BASE_URL como fallback) |
| `MOCKPAY_DB_PATH` | `mockpay.db` | Caminho do arquivo de banco de dados SQLite |
| `MOCKPAY_INTEREST_RATE` | `0` | Taxa de juros mensal padrao (%) para parcelamentos |
| `MOCKPAY_WEBHOOK_URL` | (vazio) | URL para receber eventos de webhook (vazio = desabilitado) |
| `MOCKPAY_WEBHOOK_SECRET` | (vazio) | Segredo HMAC-SHA256 para assinaturas de webhook |

As variaveis de ambiente sao carregadas via arquivo `.env` (usando `godotenv`) ou a partir do ambiente do sistema.

## Teste Rapido

```bash
# Create a billing
curl -X POST http://localhost:8080/v1/billing/create \
  -H "Authorization: Bearer mock_key" \
  -H "Content-Type: application/json" \
  -d '{
    "frequency": "ONE_TIME",
    "methods": ["PIX"],
    "products": [{"external_id": "p1", "name": "Test Product", "quantity": 1, "price": 5000}],
    "return_url": "http://localhost:3000/return",
    "completion_url": "http://localhost:3000/done"
  }'

# Open the dashboard
open http://localhost:8080/

# Create a PIX charge
curl -X POST http://localhost:8080/v1/pix/create \
  -H "Authorization: Bearer mock_key" \
  -H "Content-Type: application/json" \
  -d '{"amount": 5000, "expires_in": 3600}'

# Check health
curl http://localhost:8080/health
```

## Testes em Dispositivos Moveis

Para testar QR codes a partir de um dispositivo movel na mesma rede:

```bash
MOCKPAY_PUBLIC_URL=http://192.168.1.100:8080 go run main.go
```

Os QR codes conterao URLs utilizando o IP da sua rede local em vez de `localhost`, permitindo que dispositivos moveis acessem o servidor.

## Compilacao

```bash
go build -o mockpay .
./mockpay
```

## Executando os Testes

```bash
go test ./...
```

Os testes utilizam SQLite em memoria (`:memory:`) para isolamento.
