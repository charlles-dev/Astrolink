package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	EnvPaymentsProvider       = "PAYMENTS_PROVIDER"
	EnvMercadoPagoAccessToken = "MERCADOPAGO_ACCESS_TOKEN"
	EnvMercadoPagoAPIBaseURL  = "MERCADOPAGO_API_BASE_URL"
	EnvMercadoPagoPayerEmail  = "MERCADOPAGO_PAYER_EMAIL"
	DefaultPaymentsProvider   = "demo"
)

type Config struct {
	Env             string
	HTTPAddr        string
	NodeName        string
	AdminUser       string
	AdminPassword   string
	AdminTOTPSecret string
	JWTSecret       string
	DatabaseURL     string
	LogLevel        slog.Level

	PaymentsProvider         string
	MercadoPagoAccessToken   string
	MercadoPagoAPIBaseURL    string
	MercadoPagoPayerEmail    string
	MercadoPagoWebhookSecret string

	OpenNDSEnabled bool
	OpenNDSHost    string
	OpenNDSPort    int
	OpenNDSUser    string
	OpenNDSKeyPath string
	OpenNDSTimeout time.Duration
	OpenNDSRetries int
}

func FromEnv() Config {
	return Config{
		Env:             env("GO_ENV", "development"),
		HTTPAddr:        env("HTTP_ADDR", ":5000"),
		NodeName:        env("NODE_NAME", "dev-node-01"),
		AdminUser:       env("ADMIN_USUARIO", "admin"),
		AdminPassword:   env("ADMIN_SENHA", "admin123"),
		AdminTOTPSecret: env("ADMIN_TOTP_SECRET", ""),
		JWTSecret:       env("JWT_SECRET", "dev-jwt-secret-nao-usar-em-producao-32chars"),
		DatabaseURL:     env("DATABASE_URL", ""),
		LogLevel:        parseLogLevel(env("LOG_LEVEL", "info")),

		PaymentsProvider:         env(EnvPaymentsProvider, DefaultPaymentsProvider),
		MercadoPagoAccessToken:   env(EnvMercadoPagoAccessToken, ""),
		MercadoPagoAPIBaseURL:    env(EnvMercadoPagoAPIBaseURL, ""),
		MercadoPagoPayerEmail:    env(EnvMercadoPagoPayerEmail, ""),
		MercadoPagoWebhookSecret: env("MERCADOPAGO_WEBHOOK_SECRET", ""),

		OpenNDSEnabled: parseBool(env("OPENNDS_ENABLED", "false")),
		OpenNDSHost:    env("OPENNDS_SSH_HOST", ""),
		OpenNDSPort:    parseInt(env("OPENNDS_SSH_PORT", "22"), 22),
		OpenNDSUser:    env("OPENNDS_SSH_USER", "root"),
		OpenNDSKeyPath: env("OPENNDS_SSH_KEY_PATH", ""),
		OpenNDSTimeout: parseDuration(env("OPENNDS_SSH_TIMEOUT", "10s"), 10*time.Second),
		OpenNDSRetries: parseInt(env("OPENNDS_AUTH_RETRIES", "3"), 3),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "t", "yes", "y", "sim":
		return true
	default:
		return false
	}
}

func parseInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseDuration(value string, fallback time.Duration) time.Duration {
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
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
