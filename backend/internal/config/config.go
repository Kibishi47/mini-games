package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	AppPort            string
	FrontendURL        string
	DatabaseURL        string
	RedisAddr          string
	RedisPassword      string
	JWTSecret          string
	JWTExpirationHours int
	DiscordClientID    string
	DiscordClientSecret string
	DiscordRedirectURI string
}

func Load() *Config {
	_ = godotenv.Load()

	jwtExpHours, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "72"))

	return &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		AppPort:            getEnv("APP_PORT", "8080"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://minigames:minigames_secret@localhost:5432/minigames_db?sslmode=disable"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		JWTSecret:          getEnv("JWT_SECRET", "super_secret_jwt_key_change_in_production_min_32_chars!"),
		JWTExpirationHours: jwtExpHours,
		DiscordClientID:    getEnv("DISCORD_CLIENT_ID", ""),
		DiscordClientSecret: getEnv("DISCORD_CLIENT_SECRET", ""),
		DiscordRedirectURI: getEnv("DISCORD_REDIRECT_URI", "http://localhost:8080/api/auth/discord/callback"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
