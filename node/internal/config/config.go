package config

import (
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	Env           string
	HTTPAddr      string
	NodeName      string
	AdminUser     string
	AdminPassword string
	JWTSecret     string
	DatabaseURL   string
	LogLevel      slog.Level
}

func FromEnv() Config {
	return Config{
		Env:           env("GO_ENV", "development"),
		HTTPAddr:      env("HTTP_ADDR", ":5000"),
		NodeName:      env("NODE_NAME", "dev-node-01"),
		AdminUser:     env("ADMIN_USUARIO", "admin"),
		AdminPassword: env("ADMIN_SENHA", "admin123"),
		JWTSecret:     env("JWT_SECRET", "dev-jwt-secret-nao-usar-em-producao-32chars"),
		DatabaseURL:   env("DATABASE_URL", ""),
		LogLevel:      parseLogLevel(env("LOG_LEVEL", "info")),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func parseLogLevel(value string) slog.Level {
	switch value {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		if parsed, err := strconv.Atoi(value); err == nil {
			return slog.Level(parsed)
		}
		return slog.LevelInfo
	}
}
