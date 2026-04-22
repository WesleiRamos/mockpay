package domain

type Customer struct {
	ID        string            `json:"id"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

type CustomerRef struct {
	ID       string            `json:"id"`
	Metadata map[string]string `json:"metadata"`
}

type CustomerInput struct {
	Name      string `json:"name"`
	Cellphone string `json:"cellphone"`
	Email     string `json:"email"`
	TaxID     string `json:"tax_id"`
}
