package payments

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	ProviderDemo        = "demo"
	ProviderMercadoPago = "mercadopago"

	defaultMercadoPagoAPIBaseURL = "https://api.mercadopago.com"
	defaultMercadoPagoTimeout    = 10 * time.Second

	PixStatusPending  PixStatus = "pendente"
	PixStatusApproved PixStatus = "aprovado"
	PixStatusCanceled PixStatus = "cancelado"
	PixStatusExpired  PixStatus = "expirado"
)

var (
	ErrWebhookSecretMissing = errors.New("mercadopago webhook secret nao configurado")
	ErrWebhookInvalid       = errors.New("mercadopago webhook invalido")
)

type CreatePixInput struct {
	TXID      string
	Valor     string
	Descricao string
	ExpiresAt time.Time
}

type Pix struct {
	TXID         string
	PixCopiaCola string
	QRCodeBase64 string
	ExpiresAt    time.Time
}

type PixStatus string

type StatusQuery struct {
	PaymentID string
	TXID      string
}

type StatusResult struct {
	PaymentID string
	TXID      string
	Status    PixStatus
}

type Provider interface {
	CreatePix(context.Context, CreatePixInput) (Pix, error)
	PixStatus(context.Context, StatusQuery) (StatusResult, error)
}

type ProviderConfig struct {
	Name                   string
	MercadoPagoAccessToken string
	MercadoPagoAPIBaseURL  string
	HTTPClient             *http.Client
}

func NewProvider(cfg ProviderConfig) Provider {
	switch strings.ToLower(strings.TrimSpace(cfg.Name)) {
	case "", ProviderDemo:
		return DemoProvider{}
	case ProviderMercadoPago:
		token := strings.TrimSpace(cfg.MercadoPagoAccessToken)
		if token == "" {
			return DemoProvider{}
		}
		baseURL := strings.TrimSpace(cfg.MercadoPagoAPIBaseURL)
		if baseURL == "" {
			baseURL = defaultMercadoPagoAPIBaseURL
		}
		return MercadoPagoProvider{
			AccessToken: token,
			APIBaseURL:  strings.TrimRight(baseURL, "/"),
			HTTPClient:  mercadoPagoHTTPClient(cfg.HTTPClient),
		}
	default:
		return DemoProvider{}
	}
}

func mercadoPagoHTTPClient(client *http.Client) *http.Client {
	if client != nil {
		return client
	}
	return &http.Client{Timeout: defaultMercadoPagoTimeout}
}

type DemoProvider struct{}

func (DemoProvider) CreatePix(_ context.Context, input CreatePixInput) (Pix, error) {
	if input.TXID == "" {
		return Pix{}, fmt.Errorf("txid obrigatorio")
	}
	return Pix{
		TXID:         input.TXID,
		PixCopiaCola: "00020126580014br.gov.bcb.pix0136astrolink-demo-" + input.TXID,
		QRCodeBase64: "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNTYiIGhlaWdodD0iMjU2Ij48cmVjdCB3aWR0aD0iMjU2IiBoZWlnaHQ9IjI1NiIgZmlsbD0id2hpdGUiLz48dGV4dCB4PSIxMjgiIHk9IjEyOCIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZmlsbD0iYmxhY2siPkFzdHJvbGluayBQSVg8L3RleHQ+PC9zdmc+",
		ExpiresAt:    input.ExpiresAt,
	}, nil
}

func (DemoProvider) PixStatus(_ context.Context, input StatusQuery) (StatusResult, error) {
	return StatusResult{
		PaymentID: input.PaymentID,
		TXID:      input.TXID,
		Status:    PixStatusPending,
	}, nil
}

type MercadoPagoProvider struct {
	AccessToken string
	APIBaseURL  string
	HTTPClient  *http.Client
}

func (p MercadoPagoProvider) CreatePix(context.Context, CreatePixInput) (Pix, error) {
	return Pix{}, errors.New("mercadopago CreatePix nao implementado")
}

func (p MercadoPagoProvider) PixStatus(ctx context.Context, input StatusQuery) (StatusResult, error) {
	paymentID := strings.TrimSpace(input.PaymentID)
	if paymentID == "" {
		return StatusResult{}, fmt.Errorf("payment id obrigatorio")
	}
	endpoint := strings.TrimRight(p.APIBaseURL, "/")
	if endpoint == "" {
		endpoint = defaultMercadoPagoAPIBaseURL
	}
	endpoint += "/v1/payments/" + url.PathEscape(paymentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return StatusResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(p.AccessToken))
	req.Header.Set("Accept", "application/json")

	client := p.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return StatusResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return StatusResult{}, fmt.Errorf("mercadopago status HTTP %d", resp.StatusCode)
	}

	var body struct {
		ID                any    `json:"id"`
		Status            string `json:"status"`
		ExternalReference string `json:"external_reference"`
	}
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&body); err != nil {
		return StatusResult{}, err
	}
	resultPaymentID := stringifyPaymentID(body.ID)
	if resultPaymentID == "" {
		resultPaymentID = paymentID
	}
	txid := strings.TrimSpace(body.ExternalReference)
	if txid == "" {
		txid = input.TXID
	}
	return StatusResult{
		PaymentID: resultPaymentID,
		TXID:      txid,
		Status:    mapMercadoPagoStatus(body.Status),
	}, nil
}

func mapMercadoPagoStatus(status string) PixStatus {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "approved":
		return PixStatusApproved
	case "pending", "in_process", "in_mediation":
		return PixStatusPending
	case "cancelled", "rejected":
		return PixStatusCanceled
	case "expired":
		return PixStatusExpired
	default:
		return PixStatusPending
	}
}

func stringifyPaymentID(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return fmt.Sprintf("%.0f", typed)
	case json.Number:
		return strings.TrimSpace(typed.String())
	default:
		return ""
	}
}

type MercadoPagoWebhookVerifier struct {
	Secret string
}

type MercadoPagoWebhookVerification struct {
	DataID    string
	RequestID string
	Signature string
}

func (v MercadoPagoWebhookVerifier) Verify(input MercadoPagoWebhookVerification) error {
	secret := strings.TrimSpace(v.Secret)
	if secret == "" {
		return ErrWebhookSecretMissing
	}
	parts := parseSignature(input.Signature)
	ts := parts["ts"]
	got := parts["v1"]
	if ts == "" || got == "" || input.DataID == "" || input.RequestID == "" {
		return ErrWebhookInvalid
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte("id:" + input.DataID + ";request-id:" + input.RequestID + ";ts:" + ts + ";"))
	want := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(strings.ToLower(got)), []byte(want)) {
		return ErrWebhookInvalid
	}
	return nil
}

func parseSignature(header string) map[string]string {
	result := map[string]string{}
	for _, part := range strings.Split(header, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		result[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return result
}
