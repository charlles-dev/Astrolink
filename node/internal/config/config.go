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
	EnvAstrolinkEnvFile       = "ASTROLINK_ENV_FILE"
	EnvAstrolinkAllowEnvWrite = "ASTROLINK_ALLOW_ENV_WRITE"
	DefaultPaymentsProvider   = "demo"
)

type Config struct {
	Env                    string
	HTTPAddr               string
	AstrolinkEnvFile       string
	AstrolinkAllowEnvWrite bool
	NodeName               string
	AdminUser              string
	AdminPassword          string
	AdminTOTPSecret        string
	JWTSecret              string
	DatabaseURL            string
	LogLevel               slog.Level

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
	envFilePath := strings.TrimSpace(os.Getenv(EnvAstrolinkEnvFile))
	fileValues := loadEnvFileValues(envFilePath)
	return Config{
		Env:                    env(fileValues, "GO_ENV", "development"),
		HTTPAddr:               env(fileValues, "HTTP_ADDR", ":5000"),
		AstrolinkEnvFile:       envFilePathOrDefault(envFilePath),
		AstrolinkAllowEnvWrite: parseBool(env(fileValues, EnvAstrolinkAllowEnvWrite, "false")),
		NodeName:               env(fileValues, "NODE_NAME", "dev-node-01"),
		AdminUser:              env(fileValues, "ADMIN_USUARIO", "admin"),
		AdminPassword:          env(fileValues, "ADMIN_SENHA", "admin123"),
		AdminTOTPSecret:        env(fileValues, "ADMIN_TOTP_SECRET", ""),
		JWTSecret:              env(fileValues, "JWT_SECRET", "dev-jwt-secret-nao-usar-em-producao-32chars"),
		DatabaseURL:            env(fileValues, "DATABASE_URL", ""),
		LogLevel:               parseLogLevel(env(fileValues, "LOG_LEVEL", "info")),

		PaymentsProvider:         env(fileValues, EnvPaymentsProvider, DefaultPaymentsProvider),
		MercadoPagoAccessToken:   env(fileValues, EnvMercadoPagoAccessToken, ""),
		MercadoPagoAPIBaseURL:    env(fileValues, EnvMercadoPagoAPIBaseURL, ""),
		MercadoPagoPayerEmail:    env(fileValues, EnvMercadoPagoPayerEmail, ""),
		MercadoPagoWebhookSecret: env(fileValues, "MERCADOPAGO_WEBHOOK_SECRET", ""),

		OpenNDSEnabled: parseBool(env(fileValues, "OPENNDS_ENABLED", "false")),
		OpenNDSHost:    env(fileValues, "OPENNDS_SSH_HOST", ""),
		OpenNDSPort:    parseInt(env(fileValues, "OPENNDS_SSH_PORT", "22"), 22),
		OpenNDSUser:    env(fileValues, "OPENNDS_SSH_USER", "root"),
		OpenNDSKeyPath: env(fileValues, "OPENNDS_SSH_KEY_PATH", ""),
		OpenNDSTimeout: parseDuration(env(fileValues, "OPENNDS_SSH_TIMEOUT", "10s"), 10*time.Second),
		OpenNDSRetries: parseInt(env(fileValues, "OPENNDS_AUTH_RETRIES", "3"), 3),
	}
}

func loadEnvFileValues(path string) map[string]string {
	path = envFilePathOrDefault(path)
	file, err := LoadEnvFile(path)
	if err != nil {
		return map[string]string{}
	}
	return file.Values()
}

func envFilePathOrDefault(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ".env"
	}
	return path
}

func env(fileValues map[string]string, key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if value := fileValues[key]; value != "" {
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
