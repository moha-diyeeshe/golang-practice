package config

import (
	"log"
	"os"
	"fmt"

	"github.com/joho/godotenv"
)

// LoadEnv loads .env file (only once at startup)
func LoadEnv() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("❌ Failed to load .env:", err)
	}

	log.Println("✅ .env loaded successfully")

	// 🔥 DEBUG CHECK
	log.Println("JWT SECRET RAW:", os.Getenv("JWT_SECRET"))
}

// GetEnv returns value or empty string if not found
func GetEnv(key string) string {
	fmt.Printf("Getting env var: %s\n", key)
	return os.Getenv(key)
}
// MustGetEnv returns value or crashes if missing (for critical configs)
func MustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("❌ Missing required environment variable: %s", key)
	}
	return value
}