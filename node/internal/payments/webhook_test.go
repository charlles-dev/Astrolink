package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestMercadoPagoWebhookVerifier_AceitaAssinaturaValida(t *testing.T) {
	secret := "webhook-secret"
	dataID := "12345"
	requestID := "req-1"
	ts := "1716250000"
	signature := signTestWebhook(secret, dataID, requestID, ts)

	err := MercadoPagoWebhookVerifier{Secret: secret}.Verify(MercadoPagoWebhookVerification{
		DataID:    dataID,
		RequestID: requestID,
		Signature: "ts=" + ts + ",v1=" + signature,
	})
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
}

func TestMercadoPagoWebhookVerifier_RejeitaSemSegredo(t *testing.T) {
	err := MercadoPagoWebhookVerifier{}.Verify(MercadoPagoWebhookVerification{
		DataID:    "12345",
		RequestID: "req-1",
		Signature: "ts=1716250000,v1=abc",
	})
	if err == nil {
		t.Fatal("Verify() error = nil, want error")
	}
}

func TestDemoProvider_PixStatusPermanecePendente(t *testing.T) {
	result, err := DemoProvider{}.PixStatus(nil, StatusQuery{TXID: "ast_123"})
	if err != nil {
		t.Fatalf("PixStatus() error = %v", err)
	}
	if result.Status != PixStatusPending {
		t.Fatalf("Status = %q, want %q", result.Status, PixStatusPending)
	}
}

func signTestWebhook(secret, dataID, requestID, ts string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("id:" + dataID + ";request-id:" + requestID + ";ts:" + ts + ";"))
	return hex.EncodeToString(mac.Sum(nil))
}
