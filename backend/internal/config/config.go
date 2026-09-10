package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv        string
	AppPort       string
	FrontendURL   string
	RedisAddr     string
	RedisPassword string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		AppPort:       getEnv("APP_PORT", "8080"),
		FrontendURL:   getEnv("FRONTEND_URL", "http://localhost:3000"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
