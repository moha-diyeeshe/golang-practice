package db

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var DB *sql.DB

// ConnectDB initializes database connection
func ConnectDB() {
	dbType := os.Getenv("DB_TYPE")
	connStr := os.Getenv("DB_URL") // better naming than DB_CONN_STR

	if dbType == "" {
		log.Fatal("DB_TYPE is not set")
	}

	if connStr == "" {
		log.Fatal("DB_URL is not set")
	}

	var err error

	switch dbType {
	case "postgres":
		DB, err = sql.Open("postgres", connStr)

	case "mysql":
		DB, err = sql.Open("mysql", connStr)

	default:
		log.Fatalf("Unsupported DB_TYPE: %s", dbType)
	}

	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// 🔥 Connection pool tuning (important in real systems)
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Printf("Successfully connected to %s database", dbType)
}

// GetDB returns the global DB instance
func GetDB() *sql.DB {
	if DB == nil {
		log.Fatal("Database is not initialized. Call ConnectDB() first.")
	}
	return DB
}