# Armazenamento

O MockPay utiliza SQLite para persistência por meio do driver escrito em Go puro `modernc.org/sqlite` (sem necessidade de CGO). Os dados sobrevivem a reinicializações do servidor.

## Configuração

```bash
MOCKPAY_DB_PATH=mockpay.db   # padrão
```

Defina `MOCKPAY_DB_PATH` para personalizar o local do arquivo de banco de dados. Utilize `:memory:` para o modo em memória (útil para testes).

## Esquema do Banco de Dados

### billings

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| `id` | TEXT PK | ID hexadecimal de 32 caracteres |
| `url` | TEXT | URL de checkout |
| `amount` | INTEGER | Total em centavos |
| `status` | TEXT | PENDING, APPROVED, DENIED, CANCELLED, EXPIRED |
| `dev_mode` | INTEGER | Booleano (0/1) |
| `methods` | TEXT | Array JSON com métodos de pagamento |
| `products` | TEXT | Array JSON com objetos de produto |
| `frequency` | TEXT | ONE_TIME ou MULTIPLE_PAYMENTS |
| `next_billing` | TEXT | Data da próxima cobrança recorrente (anulável) |
| `customer` | TEXT | Referência JSON do cliente (anulável) |
| `return_url` | TEXT | URL de redirecionamento em caso de recusa |
| `completion_url` | TEXT | URL de redirecionamento em caso de aprovação |
| `installments` | INTEGER | Número de parcelas |
| `interest_rate` | REAL | Taxa de juros mensal em % |
| `installment_list` | TEXT | Array JSON com objetos de parcelas |
| `created_at` | TEXT | Timestamp de criação |
| `updated_at` | TEXT | Timestamp da última atualização |

### customers

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| `id` | TEXT PK | ID hexadecimal de 32 caracteres |
| `metadata` | TEXT | Objeto JSON com campos do cliente |
| `created_at` | TEXT | Timestamp de criação |
| `updated_at` | TEXT | Timestamp da última atualização |

### coupons

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| `id` | TEXT PK | ID hexadecimal de 32 caracteres |
| `code` | TEXT | Código único do cupom |
| `notes` | TEXT | Descrição |
| `max_redeems` | INTEGER | Número máximo de resgates |
| `redeems_count` | INTEGER | Contagem atual de resgates |
| `discount_kind` | TEXT | PERCENTAGE ou FIXED |
| `discount` | INTEGER | Valor do desconto |
| `dev_mode` | INTEGER | Booleano (0/1) |
| `status` | TEXT | ACTIVE ou INACTIVE |
| `created_at` | TEXT | Timestamp de criação |
| `updated_at` | TEXT | Timestamp da última atualização |

### pix_charges

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| `id` | TEXT PK | ID hexadecimal de 32 caracteres |
| `amount` | INTEGER | Valor em centavos |
| `status` | TEXT | PENDING, APPROVED, EXPIRED |
| `dev_mode` | INTEGER | Booleano (0/1) |
| `br_code` | TEXT | URL do QR code |
| `br_code_base64` | TEXT | Data URI em Base64 do QR code em PNG |
| `platform_fee` | INTEGER | Taxa da plataforma em centavos |
| `expires_at` | TEXT | Timestamp de expiração |
| `customer` | TEXT | Referência JSON do cliente (anulável) |
| `created_at` | TEXT | Timestamp de criação |
| `updated_at` | TEXT | Timestamp da última atualização |

### webhook_deliveries

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| `id` | TEXT PK | ID hexadecimal de 32 caracteres |
| `event_id` | TEXT | ID do evento associado |
| `url` | TEXT | URL de destino do webhook |
| `attempt` | INTEGER | Número da tentativa (1-3) |
| `status_code` | INTEGER | Código de resposta HTTP |
| `success` | INTEGER | Booleano (0/1) |
| `created_at` | TEXT | Timestamp de entrega |

## Colunas JSON

Campos complexos são serializados como texto JSON:

- `methods` → `["PIX","CARD"]`
- `products` → `[{"external_id":"p1","name":"Test","quantity":1,"price":5000}]`
- `installment_list` → `[{"number":1,"amount":1750,"status":"PENDING","due_date":"..."}]`
- `customer` → `{"id":"abc","metadata":{"email":"test@test.com"}}`
- `metadata` → `{"name":"Joao","email":"joao@test.com"}`

## Concorrência

O store utiliza `sync.RWMutex` para acesso seguro entre threads:
- Operações de escrita (`Create*`, `Update*`) adquirem um bloqueio exclusivo
- Operações de leitura (`Get*`, `List*`) adquirem um bloqueio compartilhado
- A conexão SQLite é limitada a no máximo 1 conexão aberta para segurança de escrita

## Migrações

As migrações de esquema são executadas automaticamente na inicialização por meio de `CREATE TABLE IF NOT EXISTS`. Nenhuma ferramenta de migração é necessária -- o banco de dados é criado do zero caso não exista.

## Limpeza de Dados

Para resetar todos os dados, basta excluir o arquivo do banco de dados e reiniciar:

```bash
rm mockpay.db
go run main.go
```
