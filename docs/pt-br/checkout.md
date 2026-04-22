# Checkout & Dashboard

O MockPay inclui uma página visual de checkout para cada pagamento e um dashboard para gerenciar todos os pagamentos pelo navegador. São páginas HTML servidas sem autenticação — projetadas para fluxos de desenvolvimento e testes.

## Página de Checkout

```
GET /checkout/{id}
```

Renderiza um layout em duas colunas:

### Painel Esquerdo — Resumo do Pedido
- Marca do MockPay
- Valor total (exibição em destaque)
- ID do pagamento
- Badge de status (PENDING / APPROVED / DENIED / EXPIRED)
- Método(s) de pagamento
- Lista de produtos com quantidades e preços
- Aviso "Simulated payment · testing only"

### Painel Direito — Ação de Pagamento

**Quando PENDING:**
- Abas de método (se tanto PIX quanto CARD estiverem disponíveis)
- **Painel PIX**: imagem de QR code para escaneamento
- **Painel CARD**: campos de número do cartão, nome, validade, CVV, seletor de parcelas com cálculo dinâmico
- Botão **Approve** (verde) e botão **Deny** (vermelho)

**Quando resolvido (APPROVED/DENIED/EXPIRED):**
- Ícone e mensagem de status
- Sem botões de ação

### Responsivo para Dispositivos Móveis

Em telas com largura inferior a 680px, o layout muda de duas colunas para coluna única.

## Aprovar Pagamento

```
GET /checkout/{id}/approve
```

Aprova o pagamento e redireciona:
- Para cobranças (billings): redireciona para `completion_url` (se definida), caso contrário volta para o checkout
- Para cobranças PIX: redireciona de volta para a página de checkout

## Negar Pagamento

```
GET /checkout/{id}/deny
```

Nega o pagamento e redireciona:
- Para cobranças (billings): redireciona para `return_url` (se definida), caso contrário volta para o checkout
- Para cobranças PIX: redireciona de volta para a página de checkout

## Integração com QR Code

Quando um pagamento está PENDING, a página de checkout gera um QR code real (via `go-qrcode`) contendo a URL de aprovação:

```
http://<hostname>:<port>/checkout/<id>/approve
```

Escanear este QR code com um celular abre o endpoint de aprovação diretamente, confirmando o pagamento com uma única ação.

## Dashboard

```
GET /
```

Dashboard interativo exibindo todos os pagamentos de cobranças e cobranças PIX.

### Funcionalidades

- **Cards de estatísticas** — Total, Pendentes, Aprovados, Negados/Expirados
- **Filtros clicáveis** — Clique em um card de estatísticas para filtrar a tabela por status
- **Tabela de transações** — ID, Valor, Status, Método, Tipo (billing/pix), Data de criação, Link para visualizar
- **Badges de status** — Indicadores coloridos de status para cada transação
- **Chips de tipo** — Diferencia entre os tipos de pagamento `billing` e `pix`
- **Atualização automática** — A página recarrega a cada 5 segundos com indicador de contagem regressiva
- **Animações fade-up** — Animações escalonadas nas linhas para um acabamento visual refinado

### Design

O dashboard segue o sistema de design:
- Família de fontes Inter
- Tema claro com fundo `#f7f8f5`
- Cor de destaque verde (`#9fe870`)
- Cards arredondados e botões em formato de pílula
- Barra de navegação fixa com logotipo centralizado

## Health Check

```
GET /health
```

Retorna JSON:

```json
{
  "status": "ok",
  "timestamp": "2026-04-21T12:00:00.000"
}
```

## Estatísticas

```
GET /v1/stats
```

Requer autenticação. Retorna contagens agregadas:

```json
{
  "data": {
    "billings_total": 5,
    "billings_pending": 2,
    "billings_approved": 2,
    "billings_denied": 1,
    "pix_total": 3,
    "pix_pending": 1,
    "pix_approved": 1,
    "pix_expired": 1,
    "customers_total": 4,
    "coupons_total": 2
  },
  "error": null
}
```
