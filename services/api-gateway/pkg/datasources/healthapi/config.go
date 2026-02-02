package healthapi

import (
	"os"
	"time"
)

type Config struct {
	APIBaseURL   string
	APIKey       string
	DatabaseName string
	SyncInterval time.Duration
}

func LoadConfig() *Config {
	return &Config{
		APIBaseURL:   getEnv("HEALTH_API_BASE_URL", "https://clinicaltables.nlm.nih.gov/api/conditions/v3/search"),
		APIKey:       getEnv("HEALTH_API_KEY", ""),
		DatabaseName: getEnv("HEALTH_API_DATABASE", "health_api_data"),
		SyncInterval: 1 * time.Hour,
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
