package payments

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewProvider_DefaultsToDemo(t *testing.T) {
	provider := NewProvider(ProviderConfig{})

	if _, ok := provider.(DemoProvider); !ok {
		t.Fatalf("NewProvider() = %T, want DemoProvider", provider)
	}
}

func TestNewProvider_MercadoPagoWithoutTokenFallsBackToDemo(t *testing.T) {
	provider := NewProvider(ProviderConfig{Name: ProviderMercadoPago})

	if _, ok := provider.(DemoProvider); !ok {
		t.Fatalf("NewProvider() = %T, want DemoProvider", provider)
	}
}

func TestNewProvider_MercadoPagoWithToken(t *testing.T) {
	provider := NewProvider(ProviderConfig{
		Name:                   ProviderMercadoPago,
		MercadoPagoAccessToken: "test-token",
	})

	mercadoPago, ok := provider.(MercadoPagoProvider)
	if !ok {
		t.Fatalf("NewProvider() = %T, want MercadoPagoProvider", provider)
	}
	if mercadoPago.HTTPClient == nil {
		t.Fatal("HTTPClient = nil, want default timeout client")
	}
	if mercadoPago.HTTPClient.Timeout <= 0 {
		t.Fatalf("HTTPClient.Timeout = %s, want positive timeout", mercadoPago.HTTPClient.Timeout)
	}
}

func TestMercadoPagoProvider_PixStatusFetchesPaymentAndMapsStatus(t *testing.T) {
	tests := []struct {
		name       string
		mpStatus   string
		wantStatus PixStatus
	}{
		{name: "approved", mpStatus: "approved", wantStatus: PixStatusApproved},
		{name: "pending", mpStatus: "pending", wantStatus: PixStatusPending},
		{name: "in_process", mpStatus: "in_process", wantStatus: PixStatusPending},
		{name: "in_mediation", mpStatus: "in_mediation", wantStatus: PixStatusPending},
		{name: "cancelled", mpStatus: "cancelled", wantStatus: PixStatusCanceled},
		{name: "rejected", mpStatus: "rejected", wantStatus: PixStatusCanceled},
		{name: "expired", mpStatus: "expired", wantStatus: PixStatusExpired},
		{name: "unknown", mpStatus: "charged_back", wantStatus: PixStatusPending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			var gotAuth string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotAuth = r.Header.Get("Authorization")
				_ = json.NewEncoder(w).Encode(map[string]string{
					"id":                 "12345",
					"status":             tt.mpStatus,
					"external_reference": "ast_txid",
				})
			}))
			t.Cleanup(server.Close)

			provider := MercadoPagoProvider{
				AccessToken: "test-token",
				APIBaseURL:  server.URL,
				HTTPClient:  server.Client(),
			}

			result, err := provider.PixStatus(context.Background(), StatusQuery{PaymentID: "12345"})
			if err != nil {
				t.Fatalf("PixStatus() error = %v", err)
			}
			if gotPath != "/v1/payments/12345" {
				t.Fatalf("path = %q, want /v1/payments/12345", gotPath)
			}
			if gotAuth != "Bearer test-token" {
				t.Fatalf("Authorization = %q, want Bearer test-token", gotAuth)
			}
			if result.PaymentID != "12345" {
				t.Fatalf("PaymentID = %q, want 12345", result.PaymentID)
			}
			if result.TXID != "ast_txid" {
				t.Fatalf("TXID = %q, want ast_txid", result.TXID)
			}
			if result.Status != tt.wantStatus {
				t.Fatalf("Status = %q, want %q", result.Status, tt.wantStatus)
			}
		})
	}
}
