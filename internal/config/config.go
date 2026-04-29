package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                string
	Addr               string
	LogLevel           string
	Storage            string
	DatabaseURL        string
	WebPort            string
	CORSAllowedOrigins []string
}

func Load() Config {
	_ = godotenv.Load()

	webPort := getEnv("WEB_PORT", "5173")
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))

	return Config{
		Env:                getEnv("APP_ENV", "development"),
		Addr:               getAddr(),
		LogLevel:           getEnv("APP_LOG_LEVEL", "info"),
		Storage:            getStorage(databaseURL),
		DatabaseURL:        databaseURL,
		WebPort:            webPort,
		CORSAllowedOrigins: getAllowedOrigins(webPort),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getAddr() string {
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		return ":" + port
	}

	return getEnv("APP_ADDR", ":8080")
}

func getStorage(databaseURL string) string {
	if storage := strings.TrimSpace(os.Getenv("APP_STORAGE")); storage != "" {
		return storage
	}

	if databaseURL != "" {
		return "postgres"
	}

	return "memory"
}

func getAllowedOrigins(webPort string) []string {
	configured := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if configured != "" {
		return splitAndTrim(configured)
	}

	return []string{
		fmt.Sprintf("http://localhost:%s", webPort),
		fmt.Sprintf("http://127.0.0.1:%s", webPort),
	}
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		result = append(result, part)
	}

	return result
}
