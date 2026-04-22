package i18n

var translations = map[string]map[string]string{
	// ── Dashboard ─────────────────────────────────────
	"dashboard.page_title":       {"pt-BR": "MockPay — Painel", "en": "MockPay — Dashboard"},
	"dashboard.nav_brand":        {"pt-BR": "MockPay", "en": "MockPay"},
	"dashboard.total":            {"pt-BR": "Total", "en": "Total"},
	"dashboard.pending":          {"pt-BR": "Pendente", "en": "Pending"},
	"dashboard.approved":         {"pt-BR": "Aprovado", "en": "Approved"},
	"dashboard.denied_expired":   {"pt-BR": "Negado / Expirado", "en": "Denied / Expired"},
	"dashboard.transactions":     {"pt-BR": "Transações", "en": "Transactions"},
	"dashboard.showing_all":      {"pt-BR": "Exibindo todos os pagamentos", "en": "Showing all payments"},
	"dashboard.col_id":           {"pt-BR": "ID", "en": "ID"},
	"dashboard.col_amount":       {"pt-BR": "Valor", "en": "Amount"},
	"dashboard.col_status":       {"pt-BR": "Status", "en": "Status"},
	"dashboard.col_method":       {"pt-BR": "Método", "en": "Method"},
	"dashboard.col_type":         {"pt-BR": "Tipo", "en": "Type"},
	"dashboard.col_created":      {"pt-BR": "Criado", "en": "Created"},
	"dashboard.view":             {"pt-BR": "Ver →", "en": "View →"},
	"dashboard.refreshing_in":    {"pt-BR": "Atualizando em", "en": "Refreshing in"},
	"dashboard.seconds_short":    {"pt-BR": "s", "en": "s"},
		"dashboard.auto_refresh_off": {"pt-BR": "Auto-refresh desligado", "en": "Auto-refresh off"},
	"dashboard.empty_title":      {"pt-BR": "Nenhum pagamento ainda.", "en": "No payments yet."},
	"dashboard.empty_desc":       {"pt-BR": "POST /v1/billing/create para começar", "en": "POST /v1/billing/create to get started"},
	"dashboard.lang_label":       {"pt-BR": "Idioma", "en": "Language"},

	// ── Checkout ──────────────────────────────────────
	"checkout.page_title":        {"pt-BR": "MockPay — Checkout", "en": "MockPay — Checkout"},
	"checkout.brand":             {"pt-BR": "MockPay", "en": "MockPay"},
	"checkout.total_amount":      {"pt-BR": "Valor total", "en": "Total amount"},
	"checkout.id":                {"pt-BR": "ID", "en": "ID"},
	"checkout.status":            {"pt-BR": "Status", "en": "Status"},
	"checkout.method":            {"pt-BR": "Método", "en": "Method"},
	"checkout.pix":               {"pt-BR": "PIX", "en": "PIX"},
	"checkout.card":              {"pt-BR": "Cartão", "en": "Card"},
	"checkout.items":             {"pt-BR": "Itens", "en": "Items"},
	"checkout.qty":               {"pt-BR": "qtd", "en": "qty"},
	"checkout.pay_now":           {"pt-BR": "Pagar agora", "en": "Pay now"},
	"checkout.scan_to_pay":       {"pt-BR": "Escaneie para pagar", "en": "Scan to pay"},
	"checkout.card_number":       {"pt-BR": "Número do cartão", "en": "Card number"},
	"checkout.name_on_card":      {"pt-BR": "Nome no cartão", "en": "Name on card"},
	"checkout.expiry":            {"pt-BR": "Validade", "en": "Expiry"},
	"checkout.cvv":               {"pt-BR": "CVV", "en": "CVV"},
	"checkout.installments":      {"pt-BR": "Parcelas", "en": "Installments"},
	"checkout.upfront":           {"pt-BR": "à vista", "en": "upfront"},
	"checkout.expires_at":        {"pt-BR": "Expira em", "en": "Expires at"},
	"checkout.approve":           {"pt-BR": "Aprovar pagamento", "en": "Approve payment"},
	"checkout.deny":              {"pt-BR": "Negar", "en": "Deny"},
	"checkout.approved_title":    {"pt-BR": "Pagamento aprovado", "en": "Payment approved"},
	"checkout.approved_sub":      {"pt-BR": "Esta transação foi processada com sucesso.", "en": "This transaction was successfully processed."},
	"checkout.denied_title":      {"pt-BR": "Pagamento negado", "en": "Payment denied"},
	"checkout.denied_sub":        {"pt-BR": "Esta transação foi recusada.", "en": "This transaction was refused."},
	"checkout.expired_title":     {"pt-BR": "Pagamento expirado", "en": "Payment expired"},
	"checkout.expired_sub":       {"pt-BR": "Este link de pagamento não é mais válido.", "en": "This payment link is no longer valid."},
		"checkout.discount_applied": {"pt-BR": "Desconto aplicado", "en": "Discount applied"},

	// ── Status display ────────────────────────────────
	"status.PENDING":   {"pt-BR": "PENDENTE", "en": "PENDING"},
	"status.APPROVED":  {"pt-BR": "APROVADO", "en": "APPROVED"},
	"status.DENIED":    {"pt-BR": "NEGADO", "en": "DENIED"},
	"status.EXPIRED":   {"pt-BR": "EXPIRADO", "en": "EXPIRED"},
	"status.CANCELLED": {"pt-BR": "CANCELADO", "en": "CANCELLED"},

	// ── Type display ──────────────────────────────────
	"type.billing": {"pt-BR": "cobrança", "en": "billing"},
	"type.pix":     {"pt-BR": "pix", "en": "pix"},

	// ── JS i18n ───────────────────────────────────────
	"js.showing_all":      {"pt-BR": "Exibindo todos os %d pagamentos", "en": "Showing all %d payments"},
	"js.showing_filtered": {"pt-BR": "Exibindo %d pagamento(s) %s", "en": "Showing %d %s payment(s)"},
	"js.installment_x_of": {"pt-BR": "x de ", "en": "x of "},
	"js.total":            {"pt-BR": " total", "en": " total"},
	"js.upfront":          {"pt-BR": " à vista", "en": " upfront"},

	// ── Language names ────────────────────────────────
	"lang.pt-BR": {"pt-BR": "Português", "en": "Portuguese"},
	"lang.en":    {"pt-BR": "Inglês", "en": "English"},
}

func Get(lang, key string) string {
	if m, ok := translations[key]; ok {
		if v, ok := m[lang]; ok {
			return v
		}
		if v, ok := m["pt-BR"]; ok {
			return v
		}
	}
	return key
}

func GetFunc(lang string) func(string) string {
	return func(key string) string {
		return Get(lang, key)
	}
}

func ValidLang(lang string) string {
	switch lang {
	case "pt-BR", "en":
		return lang
	default:
		return "pt-BR"
	}
}
