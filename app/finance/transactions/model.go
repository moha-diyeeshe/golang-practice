package transactions

import "time"

// Transaction is a money movement linked to a customer and the user who recorded it.
// Foreign keys: customer_id → customers.id, user_id → users.id
type Transaction struct {
	ID         int       `json:"id"`
	CustomerID int       `json:"customer_id"`
	UserID     int       `json:"user_id"`
	Amount     float64   `json:"amount"`   // For real money, consider integer cents or a decimal type
	TxnDate    time.Time `json:"txn_date"` // Business/transaction date (separate from created_at)
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateTransactionRequest is the body for creating a row (IDs come from your app / auth).
type CreateTransactionRequest struct {
	CustomerID int     `json:"customer_id"`
	UserID     int     `json:"user_id"`
	Amount     float64 `json:"amount"`
	// TxnDate is calendar date only, format YYYY-MM-DD (avoids JSON time surprises).
	TxnDate string `json:"txn_date"`
}

// UpdateTransactionRequest: only set fields you want to change (nil = leave as-is).
type UpdateTransactionRequest struct {
	CustomerID *int     `json:"customer_id,omitempty"`
	UserID     *int     `json:"user_id,omitempty"`
	Amount     *float64 `json:"amount,omitempty"`
	TxnDate    *string  `json:"txn_date,omitempty"` // YYYY-MM-DD
}

// TransactionListItem is one row in list responses: transaction + joined user and customer display fields.
type TransactionListItem struct {
	Transaction
	UserUsername  string `json:"user_username"`
	UserEmail     string `json:"user_email"`
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
}
