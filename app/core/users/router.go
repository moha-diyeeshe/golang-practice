package users

import (
	"github.com/gorilla/mux"
)

// RegisterPublicRoutes: no auth — registration, login, refresh.
func RegisterPublicRoutes(r *mux.Router, handler *UserHandler) {
	r.HandleFunc("/users", handler.CreateUser).Methods("POST")
	r.HandleFunc("/auth/login", handler.Login).Methods("POST")
	r.HandleFunc("/auth/refresh", handler.Refresh).Methods("POST")
}

// RegisterProtectedRoutes: use RequireAuth on the parent subrouter in main.
func RegisterProtectedRoutes(r *mux.Router, handler *UserHandler) {
	r.HandleFunc("/users", handler.ListUsers).Methods("GET")
	r.HandleFunc("/users/{id}", handler.GetUser).Methods("GET")
	r.HandleFunc("/users/{id}", handler.UpdateUser).Methods("PUT")
	r.HandleFunc("/users/{id}", handler.DeleteUser).Methods("DELETE")
	r.HandleFunc("/users/{id}/activate", handler.ActivateUser).Methods("PATCH")
	r.HandleFunc("/users/{id}/deactivate", handler.DeactivateUser).Methods("PATCH")
	r.HandleFunc("/auth/logout", handler.Logout).Methods("POST")
}
