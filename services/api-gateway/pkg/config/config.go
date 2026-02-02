package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	MongoDBURI string
)

func Load() error {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	MongoDBURI = os.Getenv("MONGODB_URI")
	if MongoDBURI == "" {
		log.Fatal("MONGODB_URI environment variable is required")
	}

	return nil
}

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}





