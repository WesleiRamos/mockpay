package service

import (
	"fmt"

	"github.com/wesleiramos/mockpay/internal/domain"
	"github.com/wesleiramos/mockpay/internal/store"
	"github.com/wesleiramos/mockpay/internal/util"
)

type CustomerService struct {
	store *store.MemoryStore
}

func NewCustomerService(s *store.MemoryStore) *CustomerService {
	return &CustomerService{store: s}
}

func (cs *CustomerService) Create(req domain.CustomerInput) (*domain.Customer, error) {
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}

	if existing, ok := cs.store.GetCustomerByEmail(req.Email); ok {
		return existing, nil
	}

	now := domain.NowTimestamp()
	customer := &domain.Customer{
		ID:       util.NewID(),
		Metadata: map[string]string{
			"name":      req.Name,
			"email":     req.Email,
			"cellphone": req.Cellphone,
			"tax_id":    req.TaxID,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	cs.store.CreateCustomer(customer)
	return customer, nil
}

func (cs *CustomerService) List() []*domain.Customer {
	return cs.store.ListCustomers()
}
