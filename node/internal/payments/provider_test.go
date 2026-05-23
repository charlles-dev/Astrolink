package payments

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
		MercadoPagoPayerEmail:  "cliente@example.com",
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
	if mercadoPago.PayerEmail != "cliente@example.com" {
		t.Fatalf("PayerEmail = %q, want cliente@example.com", mercadoPago.PayerEmail)
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

func TestMercadoPagoProvider_CreatePixPostsPaymentAndReturnsQRCode(t *testing.T) {
	expiresAt := time.Date(2026, 5, 22, 18, 30, 0, 0, time.UTC)
	var gotPath string
	var gotMethod string
	var gotAuth string
	var gotContentType string
	var gotIdempotencyKey string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		gotIdempotencyKey = r.Header.Get("X-Idempotency-Key")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		if err := json.Unmarshal(body, &gotBody); err != nil {
			t.Fatalf("request body JSON error = %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"external_reference": "tx-123",
			"point_of_interaction": map[string]any{
				"transaction_data": map[string]string{
					"qr_code":        "000201pix-copy-paste",
					"qr_code_base64": "base64-png",
				},
			},
		})
	}))
	t.Cleanup(server.Close)

	provider := MercadoPagoProvider{
		AccessToken: "test-token",
		APIBaseURL:  server.URL,
		HTTPClient:  server.Client(),
	}

	pix, err := provider.CreatePix(context.Background(), CreatePixInput{
		TXID:       "tx-123",
		Valor:      "42.50",
		Descricao:  "Voucher Astrolink",
		PayerEmail: "cliente@example.com",
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		t.Fatalf("CreatePix() error = %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/v1/payments" {
		t.Fatalf("path = %q, want /v1/payments", gotPath)
	}
	if gotAuth != "Bearer test-token" {
		t.Fatalf("Authorization = %q, want Bearer test-token", gotAuth)
	}
	if gotContentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", gotContentType)
	}
	if gotIdempotencyKey != "tx-123" {
		t.Fatalf("X-Idempotency-Key = %q, want tx-123", gotIdempotencyKey)
	}
	if gotBody["transaction_amount"] != 42.5 {
		t.Fatalf("transaction_amount = %#v, want 42.5", gotBody["transaction_amount"])
	}
	if gotBody["description"] != "Voucher Astrolink" {
		t.Fatalf("description = %#v, want Voucher Astrolink", gotBody["description"])
	}
	if gotBody["payment_method_id"] != "pix" {
		t.Fatalf("payment_method_id = %#v, want pix", gotBody["payment_method_id"])
	}
	if gotBody["external_reference"] != "tx-123" {
		t.Fatalf("external_reference = %#v, want tx-123", gotBody["external_reference"])
	}
	if gotBody["date_of_expiration"] != expiresAt.Format(time.RFC3339) {
		t.Fatalf("date_of_expiration = %#v, want %q", gotBody["date_of_expiration"], expiresAt.Format(time.RFC3339))
	}
	payer, ok := gotBody["payer"].(map[string]any)
	if !ok {
		t.Fatalf("payer = %#v, want object", gotBody["payer"])
	}
	if payer["email"] != "cliente@example.com" {
		t.Fatalf("payer.email = %#v, want cliente@example.com", payer["email"])
	}
	if pix.TXID != "tx-123" {
		t.Fatalf("TXID = %q, want tx-123", pix.TXID)
	}
	if pix.PixCopiaCola != "000201pix-copy-paste" {
		t.Fatalf("PixCopiaCola = %q, want QR copy/paste", pix.PixCopiaCola)
	}
	if pix.QRCodeBase64 != "base64-png" {
		t.Fatalf("QRCodeBase64 = %q, want base64-png", pix.QRCodeBase64)
	}
	if !pix.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("ExpiresAt = %s, want %s", pix.ExpiresAt, expiresAt)
	}
}

func TestMercadoPagoProvider_CreatePixUsesConfiguredPayerEmail(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("request body JSON error = %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"external_reference": "tx-123",
			"point_of_interaction": map[string]any{
				"transaction_data": map[string]string{
					"qr_code":        "000201pix-copy-paste",
					"qr_code_base64": "base64-png",
				},
			},
		})
	}))
	t.Cleanup(server.Close)

	provider := MercadoPagoProvider{
		AccessToken: "test-token",
		APIBaseURL:  server.URL,
		PayerEmail:  "configured@example.com",
		HTTPClient:  server.Client(),
	}
	if _, err := provider.CreatePix(context.Background(), CreatePixInput{
		TXID:      "tx-123",
		Valor:     "42.50",
		Descricao: "Voucher Astrolink",
	}); err != nil {
		t.Fatalf("CreatePix() error = %v", err)
	}

	payer, ok := gotBody["payer"].(map[string]any)
	if !ok {
		t.Fatalf("payer = %#v, want object", gotBody["payer"])
	}
	if payer["email"] != "configured@example.com" {
		t.Fatalf("payer.email = %#v, want configured@example.com", payer["email"])
	}
}

func TestMercadoPagoProvider_CreatePixErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{
			name:       "http error",
			statusCode: http.StatusBadRequest,
			body:       `{"message":"bad request"}`,
		},
		{
			name:       "invalid json",
			statusCode: http.StatusCreated,
			body:       `{`,
		},
		{
			name:       "missing qr code",
			statusCode: http.StatusCreated,
			body:       `{"point_of_interaction":{"transaction_data":{"qr_code_base64":"base64-png"}}}`,
		},
		{
			name:       "missing qr code base64",
			statusCode: http.StatusCreated,
			body:       `{"point_of_interaction":{"transaction_data":{"qr_code":"000201pix"}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			t.Cleanup(server.Close)

			provider := MercadoPagoProvider{
				AccessToken: "test-token",
				APIBaseURL:  server.URL,
				HTTPClient:  server.Client(),
			}

			_, err := provider.CreatePix(context.Background(), CreatePixInput{
				TXID:       "tx-123",
				Valor:      "42.50",
				Descricao:  "Voucher Astrolink",
				PayerEmail: "cliente@example.com",
			})
			if err == nil {
				t.Fatal("CreatePix() error = nil, want error")
			}
		})
	}
}
