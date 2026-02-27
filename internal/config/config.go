package config

import (
	"os"
	"time"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port            string
	DBPath          string
	SessionSecret   string
	SessionLifetime time.Duration
	AppURL          string
}

// Load reads configuration from environment variables, falling back to safe defaults.
func Load() *Config {
	return &Config{
		Port:            getenv("PORT", "8080"),
		DBPath:          getenv("DB_PATH", "./kasku.db"),
		SessionSecret:   getenv("SESSION_SECRET", "change-me-in-production-32chars!!"),
		SessionLifetime: 7 * 24 * time.Hour,
		AppURL:          getenv("APP_URL", "http://localhost:8080"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
