# PIX

Cobranças PIX representam solicitações de pagamento instantâneo. Cada cobrança gera um QR code que pode ser escaneado para aprovar o pagamento.

## Criar uma Cobrança PIX

```
POST /v1/pix/create
```

```json
{
  "amount": 5000,
  "expires_in": 3600,
  "description": "Payment description",
  "customer": {
    "name": "Joao Silva",
    "email": "joao@email.com",
    "cellphone": "11999999999",
    "tax_id": "12345678901"
  },
  "external_id": "cobranca-67890",
  "metadata": {
    "ref_pedido": "PED-2026-001"
  }
}
```

### Campos

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `amount` | int | sim | Valor em centavos (deve ser >= 100) |
| `expires_in` | int | não | Expiração em segundos (padrão: 3600 = 1 hora) |
| `description` | string | não | Descrição do pagamento |
| `customer` | object | não | Dados do cliente (cria um novo cliente) |
| `external_id` | string | não | Seu próprio identificador para esta cobrança |
| `metadata` | object | não | Pares chave-valor para dados adicionais |

### Resposta

```json
{
  "data": {
    "id": "a1b2c3d4...",
    "amount": 5000,
    "status": "PENDING",
    "dev_mode": true,
    "br_code": "http://192.168.1.100:8080/checkout/a1b2c3d4.../approve",
    "br_code_base64": "data:image/png;base64,iVBOR...",
    "platform_fee": 80,
    "expires_at": "2026-04-21T13:00:00.000",
    "customer": {...},
    "external_id": "cobranca-67890",
    "metadata": {
      "ref_pedido": "PED-2026-001"
    },
    "created_at": "2026-04-21T12:00:00.000",
    "updated_at": "2026-04-21T12:00:00.000"
  },
  "error": null
}
```

## Verificar Status do PIX

```
GET /v1/pix/check?id=<id>
```

Retorna a cobrança PIX atual com seu status.

## QR Code

Cada cobrança PIX gera um QR code real (imagem PNG, 256x256px) contendo a URL de aprovação:

```
http://<hostname>:<port>/checkout/<id>/approve
```

O QR code está disponível em dois formatos:
- `br_code` - a string da URL em texto puro
- `br_code_base64` - uma data URI em base64 (`data:image/png;base64,...`) adequada para incorporação em tags HTML `<img>`

### Configuração do Hostname

O hostname na URL do QR é controlado pela variável `MOCKPAY_PUBLIC_URL` (padrão: utiliza `MOCKPAY_BASE_URL` como fallback). Para testes em dispositivos móveis na rede local:

```bash
MOCKPAY_PUBLIC_URL=http://192.168.1.100:8080 go run main.go
```

Isso gera QR codes contendo `http://192.168.1.100:8080/checkout/.../approve`, que dispositivos móveis conseguem acessar.

Escanear o QR code com qualquer celular abre o endpoint de aprovação diretamente, confirmando o pagamento.

## Ciclo de Status

```
PENDING → APPROVED    (via aprovação no checkout ou escaneamento do QR)
PENDING → EXPIRED     (automático, após passes do expires_at)
```

## Expiração

Cobranças PIX possuem um timestamp `expires_at`. Uma tarefa em segundo plano executa a cada 30 segundos para verificar cobranças expiradas:

1. Lista todas as cobranças PIX com status `PENDING`
2. Faz o parse de cada `expires_at`
3. Se já expirada, altera o status para `EXPIRED`
4. Dispara o webhook `pix.expired`

### Expiração Padrão

Se `expires_in` não for informado, as cobranças expiram após **3600 segundos** (1 hora).

## Taxa da Plataforma

Cada cobrança PIX inclui um campo `platform_fee`, fixado em **80 centavos**. Este é um valor simulado para fins de teste.

## Fluxo de Aprovação

1. A cobrança PIX é criada com status `PENDING`
2. A página de checkout em `/checkout/<id>` exibe o QR code
3. A aprovação pode ocorrer via:
   - Clicando em "Approve payment" na página de checkout
   - Escaneando o QR code (abre a URL de aprovação)
   - Usando o dashboard em `/`
4. Na aprovação: o status muda para `APPROVED` e o webhook `pix.approved` é disparado
