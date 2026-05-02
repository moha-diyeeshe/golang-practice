package customers

import (
	"database/sql"
)

type CustomerService struct {
	DB *sql.DB
}

func NewCustomerService(db *sql.DB) *CustomerService {
	return &CustomerService{DB: db}
}

func (s *CustomerService) EnsureCustomerTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS customers (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		phone TEXT UNIQUE NOT NULL,
		address TEXT NOT NULL
	);
	`
	_, err := s.DB.Exec(query)
	return err
}

func (s *CustomerService) CreateCustomer(customer *Customer) error {
	query := `INSERT INTO customers (name, email, phone, address) 
			  VALUES ($1, $2, $3, $4) RETURNING id`
	err := s.DB.QueryRow(query, customer.Name, customer.Email, customer.Phone, customer.Address).Scan(&customer.ID)
	return err
}

func (s *CustomerService) GetCustomerByID(id int) (*Customer, error) {
	customer := &Customer{}
	query := `SELECT id, name, email, phone, address FROM customers WHERE id = $1`
	err := s.DB.QueryRow(query, id).Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Phone, &customer.Address)
	if err != nil {
		return nil, err
	}
	return customer, nil
}

func (s *CustomerService) UpdateCustomer(id int, req *UpdateCustomerRequest) (*Customer, error) {
	customer, err := s.GetCustomerByID(id)
	if err != nil {
		return nil, err
	}
	
	if req.Name != nil {
		customer.Name = *req.Name
	}
	if req.Email != nil {
		customer.Email = *req.Email
	}
	if req.Phone != nil {
		customer.Phone = *req.Phone
	}
	if req.Address != nil {
		customer.Address = *req.Address
	}
	
	query := `UPDATE customers SET name = $1, email = $2, phone = $3, address = $4 WHERE id = $5`
	_, err = s.DB.Exec(query, customer.Name, customer.Email, customer.Phone, customer.Address, id)
	if err != nil {
		return nil, err
	}
	return customer, nil
}

func (s *CustomerService) DeleteCustomer(id int) error {
	query := `DELETE FROM customers WHERE id = $1`
	_, err := s.DB.Exec(query, id)
	return err
}

func (s *CustomerService) ListCustomers() ([]*Customer, error) {
	query := `SELECT id, name, email, phone, address FROM customers`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customers := []*Customer{}
	for rows.Next() {
		customer := &Customer{}
		err := rows.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Phone, &customer.Address)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}
	return customers, nil
}



