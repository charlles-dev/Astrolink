package config

import (
	"testing"
	"time"
)

func TestFromEnv_LoadsOpenNDSConfig(t *testing.T) {
	t.Setenv("OPENNDS_ENABLED", "true")
	t.Setenv("OPENNDS_SSH_HOST", "192.168.1.1")
	t.Setenv("OPENNDS_SSH_PORT", "2222")
	t.Setenv("OPENNDS_SSH_USER", "root")
	t.Setenv("OPENNDS_SSH_KEY_PATH", "C:\\Users\\charl\\.ssh\\id_ed25519")
	t.Setenv("OPENNDS_SSH_TIMEOUT", "7s")
	t.Setenv("OPENNDS_AUTH_RETRIES", "4")

	cfg := FromEnv()

	if !cfg.OpenNDSEnabled {
		t.Fatal("OpenNDSEnabled = false, want true")
	}
	if cfg.OpenNDSHost != "192.168.1.1" {
		t.Fatalf("OpenNDSHost = %q", cfg.OpenNDSHost)
	}
	if cfg.OpenNDSPort != 2222 {
		t.Fatalf("OpenNDSPort = %d", cfg.OpenNDSPort)
	}
	if cfg.OpenNDSUser != "root" {
		t.Fatalf("OpenNDSUser = %q", cfg.OpenNDSUser)
	}
	if cfg.OpenNDSKeyPath != "C:\\Users\\charl\\.ssh\\id_ed25519" {
		t.Fatalf("OpenNDSKeyPath = %q", cfg.OpenNDSKeyPath)
	}
	if cfg.OpenNDSTimeout != 7*time.Second {
		t.Fatalf("OpenNDSTimeout = %s", cfg.OpenNDSTimeout)
	}
	if cfg.OpenNDSRetries != 4 {
		t.Fatalf("OpenNDSRetries = %d", cfg.OpenNDSRetries)
	}
}

func TestFromEnv_LoadsPaymentProviderConfig(t *testing.T) {
	t.Setenv("PAYMENTS_PROVIDER", "mercadopago")
	t.Setenv("MERCADOPAGO_ACCESS_TOKEN", "mp-token")
	t.Setenv("MERCADOPAGO_API_BASE_URL", "https://api.example.test")

	cfg := FromEnv()

	if cfg.PaymentsProvider != "mercadopago" {
		t.Fatalf("PaymentsProvider = %q", cfg.PaymentsProvider)
	}
	if cfg.MercadoPagoAccessToken != "mp-token" {
		t.Fatalf("MercadoPagoAccessToken = %q", cfg.MercadoPagoAccessToken)
	}
	if cfg.MercadoPagoAPIBaseURL != "https://api.example.test" {
		t.Fatalf("MercadoPagoAPIBaseURL = %q", cfg.MercadoPagoAPIBaseURL)
	}
}

func TestFromEnv_DefaultsPaymentProviderToDemo(t *testing.T) {
	t.Setenv("PAYMENTS_PROVIDER", "")

	cfg := FromEnv()

	if cfg.PaymentsProvider != "demo" {
		t.Fatalf("PaymentsProvider = %q, want demo", cfg.PaymentsProvider)
	}
}
