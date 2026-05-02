package main

import (
	"log"
	"net/http"

	"rest_api_go/app/core/users"
	"rest_api_go/app/crm/customers"
	"rest_api_go/app/finance/transactions"
	"rest_api_go/auth"
	"rest_api_go/config"
	"rest_api_go/db"
	"rest_api_go/middleware"

	"github.com/gorilla/mux"
)

func main() {
	config.LoadEnv()

	auth.InitRedis()
	if err := auth.PingRedis(); err != nil {
		log.Fatal("redis not reachable (start Redis or set REDIS_ADDR):", err)
	}
	log.Println("redis connected")

	db.ConnectDB()

	router := mux.NewRouter()
	router.Use(middleware.Logging)

	api := router.PathPrefix("/api/v1").Subrouter()

	userService := users.NewUserService(db.DB)
	if err := userService.EnsureUserTable(); err != nil {
		log.Fatal("failed to create table:", err)
	}
	userHandler := users.NewUserHandler(userService)

	users.RegisterPublicRoutes(api, userHandler)

	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.RequireAuth)
	users.RegisterProtectedRoutes(protected, userHandler)

	customerService := customers.NewCustomerService(db.DB)
	if err := customerService.EnsureCustomerTable(); err != nil {
		log.Fatal("failed to create customers table:", err)
	}
	customerHandler := customers.NewCustomerHandler(customerService)
	customers.RegisterProtectedRoutes(protected, customerHandler)


	transactionService := transactions.NewTransactionService(db.DB)
	if err := transactionService.EnsureTransactionTable(); err != nil {
		log.Fatal("failed to create transactions table:", err)
	}
	transactionHandler := transactions.NewTransactionHandler(transactionService)
	transactions.RegisterProtectedRoutes(protected, transactionHandler)

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
