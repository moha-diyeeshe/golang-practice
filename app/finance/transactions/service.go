package transactions

import (
	"database/sql"
	"fmt"
	"time"
)

type TransactionService struct {
	DB *sql.DB
}

func NewTransactionService(db *sql.DB) *TransactionService {
	return &TransactionService{DB: db}
}

// EnsureTransactionTable creates the table with foreign keys to customers and users.
// Call this only after customers and users tables exist (order in main matters).
func (s *TransactionService) EnsureTransactionTable() error {
	stmts := []string{
		`
		CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
			amount NUMERIC(15, 2) NOT NULL,
			txn_date DATE NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_customer_id ON transactions(customer_id);`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_txn_date ON transactions(txn_date);`,
	}
	for _, q := range stmts {
		if _, err := s.DB.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func parseTxnDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("txn_date is required (YYYY-MM-DD)")
	}
	return time.Parse(time.DateOnly, s)
}

func (s *TransactionService) Create(req *CreateTransactionRequest) (*Transaction, error) {
	txnDate, err := parseTxnDate(req.TxnDate)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	t := &Transaction{
		CustomerID: req.CustomerID,
		UserID:     req.UserID,
		Amount:     req.Amount,
		TxnDate:    txnDate,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	query := `
		INSERT INTO transactions (customer_id, user_id, amount, txn_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4::date, $5, $6)
		RETURNING id`
	err = s.DB.QueryRow(
		query,
		t.CustomerID,
		t.UserID,
		t.Amount,
		t.TxnDate,
		t.CreatedAt,
		t.UpdatedAt,
	).Scan(&t.ID)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TransactionService) GetByID(id int) (*Transaction, error) {
	t := &Transaction{}
	query := `
		SELECT id, customer_id, user_id, amount, txn_date, created_at, updated_at
		FROM transactions WHERE id = $1`
	err := s.DB.QueryRow(query, id).Scan(
		&t.ID,
		&t.CustomerID,
		&t.UserID,
		&t.Amount,
		&t.TxnDate,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TransactionService) List() ([]*TransactionListItem, error) {
	query := `
		SELECT
			t.id, t.customer_id, t.user_id, t.amount, t.txn_date, t.created_at, t.updated_at,
			u.username, u.email,
			c.name, c.email
		FROM transactions t
		INNER JOIN users u ON u.id = t.user_id
		INNER JOIN customers c ON c.id = t.customer_id
		ORDER BY t.txn_date DESC, t.id DESC`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTransactionListRows(rows)
}

// ListByCustomerID returns the same shape as List, filtered by customer.
func (s *TransactionService) ListByCustomerID(customerID int) ([]*TransactionListItem, error) {
	query := `
		SELECT
			t.id, t.customer_id, t.user_id, t.amount, t.txn_date, t.created_at, t.updated_at,
			u.username, u.email,
			c.name, c.email
		FROM transactions t
		INNER JOIN users u ON u.id = t.user_id
		INNER JOIN customers c ON c.id = t.customer_id
		WHERE t.customer_id = $1
		ORDER BY t.txn_date DESC, t.id DESC`
	rows, err := s.DB.Query(query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTransactionListRows(rows)
}

func (s *TransactionService) Update(id int, req *UpdateTransactionRequest) (*Transaction, error) {
	t, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	if req.CustomerID != nil {
		t.CustomerID = *req.CustomerID
	}
	if req.UserID != nil {
		t.UserID = *req.UserID
	}
	if req.Amount != nil {
		t.Amount = *req.Amount
	}
	if req.TxnDate != nil {
		d, err := parseTxnDate(*req.TxnDate)
		if err != nil {
			return nil, err
		}
		t.TxnDate = d
	}
	t.UpdatedAt = time.Now()

	query := `
		UPDATE transactions
		SET customer_id = $1, user_id = $2, amount = $3, txn_date = $4::date, updated_at = $5
		WHERE id = $6`
	_, err = s.DB.Exec(query, t.CustomerID, t.UserID, t.Amount, t.TxnDate, t.UpdatedAt, id)
	if err != nil {
		return nil, err
	}
	return s.GetByID(id)
}

func (s *TransactionService) Delete(id int) error {
	_, err := s.DB.Exec(`DELETE FROM transactions WHERE id = $1`, id)
	return err
}

func scanTransactionListRows(rows *sql.Rows) ([]*TransactionListItem, error) {
	var out []*TransactionListItem
	for rows.Next() {
		item := &TransactionListItem{}
		t := &item.Transaction
		err := rows.Scan(
			&t.ID,
			&t.CustomerID,
			&t.UserID,
			&t.Amount,
			&t.TxnDate,
			&t.CreatedAt,
			&t.UpdatedAt,
			&item.UserUsername,
			&item.UserEmail,
			&item.CustomerName,
			&item.CustomerEmail,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
