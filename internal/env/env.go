package env

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func GetString(key string, fallback string) string {

	if err := godotenv.Load(); err != nil {
		slog.Error("failed to load .env file", "error", err)
		os.Exit(1)
	}

	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
