package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	Port            string
	DatabaseURL     string
	AdminAPIKey     string
	AdminTenantSlug string
	AllowedProjects []string
	LogLevel        string
	IDPrefix        string
	DBQueryTimeout  time.Duration
	HTTPTimeout     time.Duration
	MaxLimit        int
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:           envOr("PORT", "8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		AdminAPIKey:    os.Getenv("API_KEY"),
		AdminTenantSlug: os.Getenv("ADMIN_TENANT_SLUG"),
		LogLevel:       envOr("LOG_LEVEL", "info"),
		IDPrefix:       envOr("ID_PREFIX", "doit"),
		DBQueryTimeout: envDuration("DB_QUERY_TIMEOUT", 10*time.Second),
		HTTPTimeout:    envDuration("HTTP_TIMEOUT", 60*time.Second),
		MaxLimit:       envInt("MAX_LIMIT", 200),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
