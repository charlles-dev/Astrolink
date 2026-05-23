package config

import "testing"

func TestSetupStatusRedactsSecrets(t *testing.T) {
	file, _ := ParseEnvFile([]byte("PAYMENTS_PROVIDER=mercadopago\nMERCADOPAGO_ACCESS_TOKEN=secret-token\nMERCADOPAGO_PAYER_EMAIL=cliente@example.com\n"))

	status := BuildSetupStatus(file)

	accessToken := setupField(t, status, "payments", EnvMercadoPagoAccessToken)
	if accessToken.Configured != true {
		t.Fatal("access token should be configured")
	}
	if accessToken.Value != "" {
		t.Fatal("secret value should not be exposed")
	}
	payerEmail := setupField(t, status, "payments", EnvMercadoPagoPayerEmail)
	if payerEmail.Value != "cliente@example.com" {
		t.Fatalf("payer email value = %q", payerEmail.Value)
	}
}

func TestApplySetupPatchRejectsUnknownKeys(t *testing.T) {
	file, _ := ParseEnvFile(nil)

	err := ApplySetupPatch(file, map[string]string{"SHELL": "powershell"})

	if err == nil {
		t.Fatal("ApplySetupPatch() error = nil, want error")
	}
}

func setupField(t *testing.T, status SetupStatus, groupKey string, fieldKey string) SetupField {
	t.Helper()

	group, ok := status.Groups[groupKey]
	if !ok {
		t.Fatalf("group %q not found", groupKey)
	}
	for _, field := range group.Fields {
		if field.Key == fieldKey {
			return field
		}
	}
	t.Fatalf("field %q not found in group %q", fieldKey, groupKey)
	return SetupField{}
}
