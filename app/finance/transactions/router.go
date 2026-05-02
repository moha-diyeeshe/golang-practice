package transactions

import (
	"github.com/gorilla/mux"
)


// RegisterProtectedRoutes registers transaction routes (mount under a router that already uses RequireAuth).
func RegisterProtectedRoutes(r *mux.Router, handler *TransactionHandler) {
	r.HandleFunc("/transactions", handler.ListTransactions).Methods("GET")
	r.HandleFunc("/transactions", handler.CreateTransaction).Methods("POST")
	r.HandleFunc("/transactions/{id}", handler.GetTransaction).Methods("GET")
	r.HandleFunc("/transactions/{id}", handler.UpdateTransaction).Methods("PUT")
}


