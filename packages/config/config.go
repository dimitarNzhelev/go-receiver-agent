package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file if it exists
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found")
	}
}

// GetEnv retrieves the value of the environment variable named by the key or returns the default value if the variable is not present
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
