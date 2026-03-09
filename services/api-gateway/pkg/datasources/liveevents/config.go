package liveevents

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseName           string
	CollectionName         string
	LocationRetentionDays  int
	AlertBufferHours       int
	EnableSearchEnrichment bool
	EnableProjections      bool
}

func LoadConfig() *Config {
	return &Config{
		DatabaseName:           getEnv("LIVE_EVENTS_DATABASE", "live_events_db"),
		CollectionName:         getEnv("LIVE_EVENTS_COLLECTION", "live_events"),
		LocationRetentionDays:  getEnvInt("LIVE_EVENTS_LOCATION_RETENTION_DAYS", 30),
		AlertBufferHours:       getEnvInt("LIVE_EVENTS_ALERT_BUFFER_HOURS", 24),
		EnableSearchEnrichment: getEnv("LIVE_EVENTS_ENABLE_SEARCH_ENRICHMENT", "false") == "true",
		EnableProjections:      getEnv("LIVE_EVENTS_ENABLE_PROJECTIONS", "false") == "true",
	}
}

func (c *Config) LocationRetention() time.Duration {
	return time.Duration(c.LocationRetentionDays) * 24 * time.Hour
}

func (c *Config) AlertBuffer() time.Duration {
	return time.Duration(c.AlertBufferHours) * time.Hour
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return n
}
