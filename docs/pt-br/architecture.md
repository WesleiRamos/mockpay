# Arquitetura

O MockPay segue uma arquitetura em camadas com clara separacao entre tratamento HTTP, logica de negocio, acesso a dados e modelos de dominio.

## Estrutura do Projeto

```
mockpay/
├── main.go                          # Ponto de entrada, injecao de dependencias e rotas
├── config/
│   └── config.go                    # Configuracao via variaveis de ambiente
├── internal/
│   ├── domain/                      # Modelos de dominio e regras de negocio
│   │   ├── response.go              # Envelope de resposta da API
│   │   ├── billing.go               # Cobranca, produtos, parcelas
│   │   ├── customer.go              # Modelos de cliente
│   │   ├── coupon.go                # Modelos de cupom
│   │   ├── pix.go                   # Modelos de cobranca PIX
│   │   └── webhook.go               # Modelos de eventos de webhook
│   ├── store/
│   │   └── memory.go                # Armazenamento baseado em SQLite
│   ├── service/                     # Camada de logica de negocio
│   │   ├── billing.go               # Cobranca + parcelas + recorrencia
│   │   ├── customer.go              # Gerenciamento de clientes
│   │   ├── coupon.go                # Gerenciamento de cupons
│   │   ├── pix.go                   # Cobrancas PIX
│   │   └── webhook.go               # Envio de webhooks + retry
│   ├── handler/                     # Handlers HTTP
│   │   ├── billing.go               # Endpoints de cobranca
│   │   ├── customer.go              # Endpoints de cliente
│   │   ├── coupon.go                # Endpoints de cupom
│   │   ├── pix.go                   # Endpoints PIX
│   │   ├── checkout.go              # Checkout + dashboard HTML
│   │   └── ui/                      # Templates HTML embarcados
│   │       ├── checkout.html
│   │       └── dashboard.html
│   ├── middleware/
│   │   └── auth.go                  # Autenticacao via Bearer token
│   └── util/
│       ├── id.go                    # Geracao de IDs aleatorios
│       └── crypto.go                # Assinaturas HMAC-SHA256
├── tests/                           # Testes de integracao e unitarios
├── docs/                            # Documentacao
├── .env.example                     # Template de variaveis de ambiente
└── go.mod
```

## Fluxo de Dependencias

```
Config
  └─→ MemoryStore (SQLite)
        ├─→ WebhookService
        │     └─→ (goroutines em segundo plano)
        ├─→ BillingService (usa WebhookService)
        ├─→ CustomerService
        ├─→ CouponService
        └─→ PixService (usa WebhookService)
              │
              ├─→ Handlers (Billing, Customer, Coupon, Pix)
              └─→ CheckoutHandler (usa Store + servicos Billing + Pix)
                    │
                    └─→ Rotas do Fiber
```

## Camadas

### Domain (`internal/domain/`)

Modelos de dados puros sem dependencias de frameworks ou pacotes externos. Contem:
- Structs de entidades (Billing, PixCharge, Customer, Coupon)
- Enums de status e constantes
- Funcoes de calculo de negocio (`CalculateInstallments`, `NowTimestamp`)
- DTOs de request/response

### Store (`internal/store/`)

Camada de persistencia de dados usando SQLite via `modernc.org/sqlite` (Go puro, sem CGO). Campos complexos (arrays, maps) sao serializados como colunas de texto JSON. Todos os metodos sao thread-safe via `sync.RWMutex`.

### Service (`internal/service/`)

Camada de logica de negocio. Cada servico encapsula validacao, transicoes de estado e efeitos colaterais (webhooks). Os servicos dependem do store e, opcionalmente, uns dos outros (por exemplo, BillingService depende de WebhookService).

### Handler (`internal/handler/`)

Camada HTTP usando Fiber v3. Os handlers de API retornam JSON. O CheckoutHandler renderiza templates HTML usando `go:embed` e `html/template`.

### Middleware (`internal/middleware/`)

O middleware `Auth` valida o cabecalho `Authorization: Bearer <token>` contra a API key configurada. Aplicado a todas as rotas `/v1/*`. Os endpoints de checkout, dashboard e health sao publicos.

## Jobs em Segundo Plano

Uma unica goroutine com um ticker de 30 segundos gerencia:

1. **Expiracao PIX** - Marca cobrancas PIX pendentes que ultrapassaram seu `expires_at` como `EXPIRED`, despachando o webhook `pix.expired`.
2. **Cobranca Recorrente** - Escaneia cobrancas `MULTIPLE_PAYMENTS` com `next_billing` no passado, criando automaticamente novas cobrancas PENDING a partir do mesmo template e despachando o webhook `billing.created`.

## IDs

Todos os IDs de entidades sao strings hexadecimais de 32 caracteres geradas a partir de 16 bytes criptograficamente aleatorios. Sem prefixos ou padroes sequenciais.

## Valores Monetarios

Todos os valores sao inteiros em **centavos** (Real brasileiro). Por exemplo, `R$ 50,00` e representado como `5000`. Isso evita problemas de precisao com ponto flutuante.

## Formato de Resposta

Toda resposta da API segue um envelope padrao:

**Sucesso:**
```json
{
  "data": { ... },
  "error": null
}
```

**Erro:**
```json
{
  "data": null,
  "error": {
    "message": "Error description",
    "code": "ERROR_CODE"
  }
}
```
