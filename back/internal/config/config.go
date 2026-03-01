package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                string
	Port               int
	CorsAllowedOrigins []string
	DatabaseURL        string
	AuthTokenTTL       time.Duration
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		Env:                getEnv("APP_ENV", "local"),
		Port:               getEnvAsInt("PORT", 8080),
		CorsAllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "")),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		AuthTokenTTL:       getEnvAsDuration("AUTH_TOKEN_TTL", time.Hour*24*7),
	}

	if cfg.DatabaseURL == "" {
		return cfg, errors.New("the DATABASE_URL environment variable is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	v := getEnv(key, "")
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	v := getEnv(key, "")
	if v == "" {
		return fallback
	}

	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}

	return d
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
