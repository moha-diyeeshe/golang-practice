package customers

import (
	"github.com/gorilla/mux"
)

// RegisterProtectedRoutes registers customer routes (mount under a router that already uses RequireAuth).
func RegisterProtectedRoutes(r *mux.Router, handler *CustomerHandler) {
	r.HandleFunc("/customers", handler.CreateCustomer).Methods("POST")
	r.HandleFunc("/customers", handler.ListCustomers).Methods("GET")
	r.HandleFunc("/customers/{id}", handler.GetCustomer).Methods("GET")
	r.HandleFunc("/customers/{id}", handler.UpdateCustomer).Methods("PUT")
	r.HandleFunc("/customers/{id}", handler.DeleteCustomer).Methods("DELETE")
}
