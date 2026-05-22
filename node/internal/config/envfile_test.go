package config

import (
	"strings"
	"testing"
)

func TestParseEnvFilePreservesCommentsAndUpdatesValues(t *testing.T) {
	input := []byte("# Astrolink\nPAYMENTS_PROVIDER=demo\nMERCADOPAGO_ACCESS_TOKEN=\"old-token\"\n\nUNKNOWN=value\n")

	file, err := ParseEnvFile(input)
	if err != nil {
		t.Fatalf("ParseEnvFile() error = %v", err)
	}
	file.Set("MERCADOPAGO_ACCESS_TOKEN", "new token")
	file.Set("MERCADOPAGO_PAYER_EMAIL", "cliente@example.com")

	got := string(file.Bytes())
	for _, want := range []string{
		"# Astrolink",
		"PAYMENTS_PROVIDER=demo",
		"MERCADOPAGO_ACCESS_TOKEN=\"new token\"",
		"UNKNOWN=value",
		"MERCADOPAGO_PAYER_EMAIL=cliente@example.com",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}

func TestParseEnvFileReadsBlankAndUnquotedValues(t *testing.T) {
	file, err := ParseEnvFile([]byte("ADMIN_TOTP_SECRET=\nOPENNDS_ENABLED=false\n"))
	if err != nil {
		t.Fatalf("ParseEnvFile() error = %v", err)
	}
	if got := file.Get("ADMIN_TOTP_SECRET"); got != "" {
		t.Fatalf("ADMIN_TOTP_SECRET = %q, want empty", got)
	}
	if got := file.Get("OPENNDS_ENABLED"); got != "false" {
		t.Fatalf("OPENNDS_ENABLED = %q, want false", got)
	}
}

func TestParseEnvFileDoesNotGrowTrailingBlankLines(t *testing.T) {
	file, err := ParseEnvFile([]byte("PAYMENTS_PROVIDER=demo\n"))
	if err != nil {
		t.Fatalf("ParseEnvFile() error = %v", err)
	}
	file.Set("MERCADOPAGO_PAYER_EMAIL", "cliente@example.com")

	got := string(file.Bytes())
	want := "PAYMENTS_PROVIDER=demo\nMERCADOPAGO_PAYER_EMAIL=cliente@example.com\n"
	if got != want {
		t.Fatalf("Bytes() = %q, want %q", got, want)
	}
}

func TestEnvFileRejectsMultilineValues(t *testing.T) {
	file, _ := ParseEnvFile(nil)
	file.Set("MERCADOPAGO_PAYER_EMAIL", "cliente@example.com\nSHELL=powershell")

	got := string(file.Bytes())

	if strings.Contains(got, "\nSHELL=") {
		t.Fatalf("Bytes() allowed newline injection:\n%s", got)
	}
	if !strings.Contains(got, `MERCADOPAGO_PAYER_EMAIL="cliente@example.com SHELL=powershell"`) {
		t.Fatalf("Bytes() did not normalize newline:\n%s", got)
	}
}
