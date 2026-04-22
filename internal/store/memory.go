package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	_ "modernc.org/sqlite"

	"github.com/wesleiramos/mockpay/internal/domain"
)

type MemoryStore struct {
	mu sync.RWMutex
	db *sql.DB
}

func New() *MemoryStore {
	return NewWithPath("mockpay.db")
}

func NewWithPath(dbPath string) *MemoryStore {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open sqlite: %v", err)
	}

	db.SetMaxOpenConns(1)

	s := &MemoryStore{db: db}
	s.migrate()
	return s
}

func (s *MemoryStore) Close() {
	s.db.Close()
}

func (s *MemoryStore) migrate() {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS billings (
			id TEXT PRIMARY KEY,
			url TEXT DEFAULT '',
			amount INTEGER NOT NULL,
			original_amount INTEGER DEFAULT 0,
			coupon_code TEXT DEFAULT '',
			status TEXT NOT NULL,
			dev_mode INTEGER DEFAULT 1,
			methods TEXT DEFAULT '[]',
			products TEXT DEFAULT '[]',
			frequency TEXT DEFAULT 'ONE_TIME',
			next_billing TEXT,
			customer TEXT,
			return_url TEXT DEFAULT '',
			completion_url TEXT DEFAULT '',
			installments INTEGER DEFAULT 1,
			interest_rate REAL DEFAULT 0,
			installment_list TEXT DEFAULT '[]',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS customers (
			id TEXT PRIMARY KEY,
			metadata TEXT DEFAULT '{}',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS coupons (
			id TEXT PRIMARY KEY,
			code TEXT NOT NULL,
			notes TEXT DEFAULT '',
			max_redeems INTEGER DEFAULT 0,
			redeems_count INTEGER DEFAULT 0,
			discount_kind TEXT NOT NULL,
			discount INTEGER NOT NULL,
			dev_mode INTEGER DEFAULT 1,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS pix_charges (
			id TEXT PRIMARY KEY,
			amount INTEGER NOT NULL,
			status TEXT NOT NULL,
			dev_mode INTEGER DEFAULT 1,
			br_code TEXT DEFAULT '',
			br_code_base64 TEXT DEFAULT '',
			platform_fee INTEGER DEFAULT 0,
			expires_at TEXT DEFAULT '',
			customer TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS webhook_deliveries (
			id TEXT PRIMARY KEY,
			event_id TEXT NOT NULL,
			url TEXT DEFAULT '',
			attempt INTEGER DEFAULT 0,
			status_code INTEGER DEFAULT 0,
			success INTEGER DEFAULT 0,
			created_at TEXT NOT NULL
		)`,
	}

	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
	}
}

// --- Billing ---

func (s *MemoryStore) CreateBilling(b *domain.Billing) {
	s.mu.Lock()
	defer s.mu.Unlock()

	methods, _ := json.Marshal(b.Methods)
	products, _ := json.Marshal(b.Products)
	installments, _ := json.Marshal(b.InstallmentList)

	var customer *string
	if b.Customer != nil {
		j, _ := json.Marshal(b.Customer)
		s := string(j)
		customer = &s
	}

	var nextBilling *string
	if b.NextBilling != nil {
		nextBilling = b.NextBilling
	}

	_, err := s.db.Exec(`INSERT INTO billings
		(id, url, amount, original_amount, coupon_code, status, dev_mode, methods, products, frequency, next_billing, customer, return_url, completion_url, installments, interest_rate, installment_list, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.ID, b.URL, b.Amount, b.OriginalAmount, b.CouponCode, string(b.Status), boolToInt(b.DevMode), string(methods), string(products),
		string(b.Frequency), nextBilling, customer, b.ReturnURL, b.CompletionURL,
		b.Installments, b.InterestRate, string(installments), b.CreatedAt, b.UpdatedAt)
	if err != nil {
		log.Printf("CreateBilling error: %v", err)
	}
}

func (s *MemoryStore) GetBilling(id string) (*domain.Billing, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(`SELECT id, url, amount, original_amount, coupon_code, status, dev_mode, methods, products, frequency, next_billing, customer,
		return_url, completion_url, installments, interest_rate, installment_list, created_at, updated_at
		FROM billings WHERE id = ?`, id)

	return s.scanBilling(row)
}

func (s *MemoryStore) ListBillings() []*domain.Billing {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT id, url, amount, original_amount, coupon_code, status, dev_mode, methods, products, frequency, next_billing, customer,
		return_url, completion_url, installments, interest_rate, installment_list, created_at, updated_at
		FROM billings ORDER BY created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []*domain.Billing
	for rows.Next() {
		b, ok := s.scanBillingRow(rows)
		if ok {
			result = append(result, b)
		}
	}
	return result
}

func (s *MemoryStore) UpdateBillingStatus(id string, status domain.BillingStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := domain.NowTimestamp()
	_, err := s.db.Exec(`UPDATE billings SET status = ?, updated_at = ? WHERE id = ?`, string(status), now, id)
	if err != nil {
		log.Printf("UpdateBillingStatus error: %v", err)
	}
}

func (s *MemoryStore) ListBillingsByStatus(status domain.BillingStatus) []*domain.Billing {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT id, url, amount, original_amount, coupon_code, status, dev_mode, methods, products, frequency, next_billing, customer,
		return_url, completion_url, installments, interest_rate, installment_list, created_at, updated_at
		FROM billings WHERE status = ? ORDER BY created_at DESC`, string(status))
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []*domain.Billing
	for rows.Next() {
		b, ok := s.scanBillingRow(rows)
		if ok {
			result = append(result, b)
		}
	}
	return result
}

func (s *MemoryStore) UpdateBillingInstallments(id string, installments []domain.Installment) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, _ := json.Marshal(installments)
	now := domain.NowTimestamp()
	_, err := s.db.Exec(`UPDATE billings SET installment_list = ?, updated_at = ? WHERE id = ?`, string(data), now, id)
	if err != nil {
		log.Printf("UpdateBillingInstallments error: %v", err)
	}
}

func (s *MemoryStore) scanBilling(row *sql.Row) (*domain.Billing, bool) {
	var b domain.Billing
	var status, frequency string
	var devMode int
	var methodsJSON, productsJSON, installmentsJSON string
	var customerJSON sql.NullString
	var nextBilling sql.NullString
	var couponCode sql.NullString

	err := row.Scan(&b.ID, &b.URL, &b.Amount, &b.OriginalAmount, &couponCode, &status, &devMode, &methodsJSON, &productsJSON, &frequency,
		&nextBilling, &customerJSON, &b.ReturnURL, &b.CompletionURL,
		&b.Installments, &b.InterestRate, &installmentsJSON, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, false
	}

	b.Status = domain.BillingStatus(status)
	b.DevMode = devMode == 1
	b.Frequency = domain.BillingFrequency(frequency)
	if couponCode.Valid {
		b.CouponCode = couponCode.String
	}

	json.Unmarshal([]byte(methodsJSON), &b.Methods)
	json.Unmarshal([]byte(productsJSON), &b.Products)
	json.Unmarshal([]byte(installmentsJSON), &b.InstallmentList)

	if customerJSON.Valid {
		var ref domain.CustomerRef
		json.Unmarshal([]byte(customerJSON.String), &ref)
		b.Customer = &ref
	}
	if nextBilling.Valid {
		b.NextBilling = &nextBilling.String
	}

	return &b, true
}

func (s *MemoryStore) scanBillingRow(rows *sql.Rows) (*domain.Billing, bool) {
	var b domain.Billing
	var status, frequency string
	var devMode int
	var methodsJSON, productsJSON, installmentsJSON string
	var customerJSON sql.NullString
	var nextBilling sql.NullString
	var couponCode sql.NullString

	err := rows.Scan(&b.ID, &b.URL, &b.Amount, &b.OriginalAmount, &couponCode, &status, &devMode, &methodsJSON, &productsJSON, &frequency,
		&nextBilling, &customerJSON, &b.ReturnURL, &b.CompletionURL,
		&b.Installments, &b.InterestRate, &installmentsJSON, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, false
	}

	b.Status = domain.BillingStatus(status)
	b.DevMode = devMode == 1
	b.Frequency = domain.BillingFrequency(frequency)
	if couponCode.Valid {
		b.CouponCode = couponCode.String
	}

	json.Unmarshal([]byte(methodsJSON), &b.Methods)
	json.Unmarshal([]byte(productsJSON), &b.Products)
	json.Unmarshal([]byte(installmentsJSON), &b.InstallmentList)

	if customerJSON.Valid {
		var ref domain.CustomerRef
		json.Unmarshal([]byte(customerJSON.String), &ref)
		b.Customer = &ref
	}
	if nextBilling.Valid {
		b.NextBilling = &nextBilling.String
	}

	return &b, true
}

// --- Customer ---

func (s *MemoryStore) CreateCustomer(c *domain.Customer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	metadata, _ := json.Marshal(c.Metadata)
	_, err := s.db.Exec(`INSERT INTO customers (id, metadata, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		c.ID, string(metadata), c.CreatedAt, c.UpdatedAt)
	if err != nil {
		log.Printf("CreateCustomer error: %v", err)
	}
}

func (s *MemoryStore) GetCustomer(id string) (*domain.Customer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(`SELECT id, metadata, created_at, updated_at FROM customers WHERE id = ?`, id)
	var c domain.Customer
	var metadataJSON string
	err := row.Scan(&c.ID, &metadataJSON, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, false
	}
	json.Unmarshal([]byte(metadataJSON), &c.Metadata)
	return &c, true
}

func (s *MemoryStore) GetCustomerByEmail(email string) (*domain.Customer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(`SELECT id, metadata, created_at, updated_at FROM customers WHERE json_extract(metadata, '$.email') = ?`, email)
	var c domain.Customer
	var metadataJSON string
	err := row.Scan(&c.ID, &metadataJSON, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, false
	}
	json.Unmarshal([]byte(metadataJSON), &c.Metadata)
	return &c, true
}

func (s *MemoryStore) ListCustomers() []*domain.Customer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT id, metadata, created_at, updated_at FROM customers ORDER BY created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []*domain.Customer
	for rows.Next() {
		var c domain.Customer
		var metadataJSON string
		if err := rows.Scan(&c.ID, &metadataJSON, &c.CreatedAt, &c.UpdatedAt); err != nil {
			continue
		}
		json.Unmarshal([]byte(metadataJSON), &c.Metadata)
		result = append(result, &c)
	}
	return result
}

// --- Coupon ---

func (s *MemoryStore) CreateCoupon(c *domain.Coupon) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`INSERT INTO coupons (id, code, notes, max_redeems, redeems_count, discount_kind, discount, dev_mode, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.Code, c.Notes, c.MaxRedeems, c.RedeemsCount, string(c.DiscountKind), c.Discount,
		boolToInt(c.DevMode), string(c.Status), c.CreatedAt, c.UpdatedAt)
	if err != nil {
		log.Printf("CreateCoupon error: %v", err)
	}
}

func (s *MemoryStore) GetCoupon(id string) (*domain.Coupon, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(`SELECT id, code, notes, max_redeems, redeems_count, discount_kind, discount, dev_mode, status, created_at, updated_at
		FROM coupons WHERE id = ?`, id)
	return s.scanCoupon(row)
}

func (s *MemoryStore) GetCouponByCode(code string) (*domain.Coupon, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(`SELECT id, code, notes, max_redeems, redeems_count, discount_kind, discount, dev_mode, status, created_at, updated_at
		FROM coupons WHERE code = ?`, code)
	return s.scanCoupon(row)
}

func (s *MemoryStore) ListCoupons() []*domain.Coupon {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT id, code, notes, max_redeems, redeems_count, discount_kind, discount, dev_mode, status, created_at, updated_at
		FROM coupons ORDER BY created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []*domain.Coupon
	for rows.Next() {
		c, ok := s.scanCouponRow(rows)
		if ok {
			result = append(result, c)
		}
	}
	return result
}

func (s *MemoryStore) IncrementCouponRedeems(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := domain.NowTimestamp()
	_, err := s.db.Exec(`UPDATE coupons SET redeems_count = redeems_count + 1, updated_at = ? WHERE id = ?`, now, id)
	if err != nil {
		log.Printf("IncrementCouponRedeems error: %v", err)
	}
}

func (s *MemoryStore) scanCoupon(row *sql.Row) (*domain.Coupon, bool) {
	var c domain.Coupon
	var discountKind, status string
	var devMode int
	err := row.Scan(&c.ID, &c.Code, &c.Notes, &c.MaxRedeems, &c.RedeemsCount, &discountKind, &c.Discount, &devMode, &status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, false
	}
	c.DiscountKind = domain.DiscountKind(discountKind)
	c.Status = domain.CouponStatus(status)
	c.DevMode = devMode == 1
	return &c, true
}

func (s *MemoryStore) scanCouponRow(rows *sql.Rows) (*domain.Coupon, bool) {
	var c domain.Coupon
	var discountKind, status string
	var devMode int
	err := rows.Scan(&c.ID, &c.Code, &c.Notes, &c.MaxRedeems, &c.RedeemsCount, &discountKind, &c.Discount, &devMode, &status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, false
	}
	c.DiscountKind = domain.DiscountKind(discountKind)
	c.Status = domain.CouponStatus(status)
	c.DevMode = devMode == 1
	return &c, true
}

// --- PIX ---

func (s *MemoryStore) CreatePixCharge(p *domain.PixCharge) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var customer *string
	if p.Customer != nil {
		j, _ := json.Marshal(p.Customer)
		s := string(j)
		customer = &s
	}

	_, err := s.db.Exec(`INSERT INTO pix_charges (id, amount, status, dev_mode, br_code, br_code_base64, platform_fee, expires_at, customer, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Amount, string(p.Status), boolToInt(p.DevMode), p.BrCode, p.BrCodeBase64,
		p.PlatformFee, p.ExpiresAt, customer, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		log.Printf("CreatePixCharge error: %v", err)
	}
}

func (s *MemoryStore) GetPixCharge(id string) (*domain.PixCharge, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(`SELECT id, amount, status, dev_mode, br_code, br_code_base64, platform_fee, expires_at, customer, created_at, updated_at
		FROM pix_charges WHERE id = ?`, id)
	return s.scanPixCharge(row)
}

func (s *MemoryStore) ListPixCharges() []*domain.PixCharge {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT id, amount, status, dev_mode, br_code, br_code_base64, platform_fee, expires_at, customer, created_at, updated_at
		FROM pix_charges ORDER BY created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []*domain.PixCharge
	for rows.Next() {
		p, ok := s.scanPixChargeRow(rows)
		if ok {
			result = append(result, p)
		}
	}
	return result
}

func (s *MemoryStore) UpdatePixStatus(id string, status domain.PixStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := domain.NowTimestamp()
	_, err := s.db.Exec(`UPDATE pix_charges SET status = ?, updated_at = ? WHERE id = ?`, string(status), now, id)
	if err != nil {
		log.Printf("UpdatePixStatus error: %v", err)
	}
}

func (s *MemoryStore) scanPixCharge(row *sql.Row) (*domain.PixCharge, bool) {
	var p domain.PixCharge
	var status string
	var devMode int
	var customerJSON sql.NullString

	err := row.Scan(&p.ID, &p.Amount, &status, &devMode, &p.BrCode, &p.BrCodeBase64,
		&p.PlatformFee, &p.ExpiresAt, &customerJSON, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, false
	}

	p.Status = domain.PixStatus(status)
	p.DevMode = devMode == 1

	if customerJSON.Valid {
		var ref domain.CustomerRef
		json.Unmarshal([]byte(customerJSON.String), &ref)
		p.Customer = &ref
	}
	return &p, true
}

func (s *MemoryStore) scanPixChargeRow(rows *sql.Rows) (*domain.PixCharge, bool) {
	var p domain.PixCharge
	var status string
	var devMode int
	var customerJSON sql.NullString

	err := rows.Scan(&p.ID, &p.Amount, &status, &devMode, &p.BrCode, &p.BrCodeBase64,
		&p.PlatformFee, &p.ExpiresAt, &customerJSON, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, false
	}

	p.Status = domain.PixStatus(status)
	p.DevMode = devMode == 1

	if customerJSON.Valid {
		var ref domain.CustomerRef
		json.Unmarshal([]byte(customerJSON.String), &ref)
		p.Customer = &ref
	}
	return &p, true
}

// --- Webhook Deliveries ---

func (s *MemoryStore) CreateDelivery(d *domain.WebhookDelivery) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`INSERT INTO webhook_deliveries (id, event_id, url, attempt, status_code, success, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.EventID, d.URL, d.Attempt, d.StatusCode, boolToInt(d.Success), d.CreatedAt)
	if err != nil {
		log.Printf("CreateDelivery error: %v", err)
	}
}

func (s *MemoryStore) UpdateDelivery(d *domain.WebhookDelivery) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`UPDATE webhook_deliveries SET event_id = ?, url = ?, attempt = ?, status_code = ?, success = ? WHERE id = ?`,
		d.EventID, d.URL, d.Attempt, d.StatusCode, boolToInt(d.Success), d.ID)
	if err != nil {
		log.Printf("UpdateDelivery error: %v", err)
	}
}

// --- Stats ---

func (s *MemoryStore) Stats() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]int64{}

	row := s.db.QueryRow(`SELECT
		COUNT(*) as total,
		COALESCE(SUM(CASE WHEN status = 'PENDING' THEN 1 ELSE 0 END), 0) as pending,
		COALESCE(SUM(CASE WHEN status = 'APPROVED' THEN 1 ELSE 0 END), 0) as approved,
		COALESCE(SUM(CASE WHEN status = 'DENIED' THEN 1 ELSE 0 END), 0) as denied
		FROM billings`)
	var total, pending, approved, denied int64
	row.Scan(&total, &pending, &approved, &denied)
	stats["billings_total"] = total
	stats["billings_pending"] = pending
	stats["billings_approved"] = approved
	stats["billings_denied"] = denied

	row = s.db.QueryRow(`SELECT
		COUNT(*) as total,
		COALESCE(SUM(CASE WHEN status = 'PENDING' THEN 1 ELSE 0 END), 0) as pending,
		COALESCE(SUM(CASE WHEN status = 'APPROVED' THEN 1 ELSE 0 END), 0) as approved,
		COALESCE(SUM(CASE WHEN status = 'EXPIRED' THEN 1 ELSE 0 END), 0) as expired
		FROM pix_charges`)
	var ptotal, ppending, papproved, pexpired int64
	row.Scan(&ptotal, &ppending, &papproved, &pexpired)
	stats["pix_total"] = ptotal
	stats["pix_pending"] = ppending
	stats["pix_approved"] = papproved
	stats["pix_expired"] = pexpired

	var ccount int64
	s.db.QueryRow(`SELECT COUNT(*) FROM customers`).Scan(&ccount)
	stats["customers_total"] = ccount

	var cpcount int64
	s.db.QueryRow(`SELECT COUNT(*) FROM coupons`).Scan(&cpcount)
	stats["coupons_total"] = cpcount

	return stats
}

// --- All Payments ---

type PaymentEntry struct {
	ID        string
	Amount    int64
	Status    string
	Method    string
	Type      string
	CreatedAt string
}

func (s *MemoryStore) ListAllPayments() []*PaymentEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT id, amount, status, methods, type, created_at FROM (
			SELECT id, amount, status, methods, 'billing' as type, created_at FROM billings
			UNION ALL
			SELECT id, amount, status, '["PIX"]' as methods, 'pix' as type, created_at FROM pix_charges
		) ORDER BY created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []*PaymentEntry
	for rows.Next() {
		var e PaymentEntry
		var methodsJSON string
		if err := rows.Scan(&e.ID, &e.Amount, &e.Status, &methodsJSON, &e.Type, &e.CreatedAt); err != nil {
			continue
		}
		var methods []string
		json.Unmarshal([]byte(methodsJSON), &methods)
		if len(methods) > 0 {
			e.Method = methods[0]
		}
		e.Status = string(e.Status)
		result = append(result, &e)
	}
	return result
}

// --- FindByID ---

func (s *MemoryStore) FindByID(id string) (any, bool) {
	if b, ok := s.GetBilling(id); ok {
		return b, true
	}
	if p, ok := s.GetPixCharge(id); ok {
		return p, true
	}
	return nil, false
}

// --- helpers ---

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Open returns a formatted connection string for testing.
func Open(dsn string) (*sql.DB, error) {
	return sql.Open("sqlite", dsn)
}

func init() {
	_ = fmt.Sprintf // suppress unused import
}
