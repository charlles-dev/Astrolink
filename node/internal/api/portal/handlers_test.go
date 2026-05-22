package portal_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/astrolink/node/internal/api/portal"
	"github.com/astrolink/node/internal/domain/planos"
	"github.com/astrolink/node/internal/gateway"
	"github.com/astrolink/node/internal/infra/memory"
	"github.com/astrolink/node/internal/payments"
	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

func TestResgatarVoucher_CodigoInexistente_Retorna404(t *testing.T) {
	app := fiber.New()
	portal.Register(app, portal.Dependencies{Store: &fakeStore{redeemErr: store.ErrVoucherNotFound}})

	req := httptest.NewRequest("POST", "/api/voucher/resgatar", strings.NewReader(`{
		"codigo": "XXXX-9999",
		"mac": "AA:BB:CC:DD:EE:FF",
		"ip": "192.168.1.50"
	}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 404 {
		t.Fatalf("esperava status 404, got %d body=%s", resp.StatusCode, string(body))
	}
	if !strings.Contains(string(body), `"erro":"nao_encontrado"`) {
		t.Fatalf("esperava erro nao_encontrado, got %s", string(body))
	}
}

func TestResgatarVoucher_Sucesso_RetornaContratoDoPortal(t *testing.T) {
	app := fiber.New()
	router := &fakeGateway{}
	duracao := 1440
	fim := time.Date(2026, 5, 22, 6, 34, 0, 0, time.UTC)
	appStore := &fakeStore{
		redeemResult: store.RedeemVoucherResult{
			Usuario: store.Usuario{
				MAC:                   "AA:BB:CC:DD:EE:FF",
				Status:                "ativo",
				FimAcesso:             &fim,
				TempoRestanteSegundos: 86400,
			},
			Plano: planos.Plano{
				ID:             2,
				Nome:           "Acesso 24 Horas",
				DuracaoMinutos: &duracao,
			},
			HadAccess: false,
		},
	}
	portal.Register(app, portal.Dependencies{Store: appStore, Gateway: router})

	req := httptest.NewRequest("POST", "/api/voucher/resgatar", strings.NewReader(`{
		"codigo": "TEST-1234",
		"mac": "AA:BB:CC:DD:EE:FF",
		"ip": "192.168.1.50"
	}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Fatalf("esperava status 200, got %d body=%s", resp.StatusCode, string(body))
	}
	for _, want := range []string{
		`"sucesso":true`,
		`"plano":"Acesso 24 Horas"`,
		`"tempo_adicionado_minutos":1440`,
		`"roteador_autorizado":true`,
	} {
		if !strings.Contains(string(body), want) {
			t.Fatalf("resposta nao contem %s: %s", want, string(body))
		}
	}
	if len(router.authorizations) != 1 {
		t.Fatalf("authorizations len = %d, want 1", len(router.authorizations))
	}
	got := router.authorizations[0]
	if got.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Fatalf("authorized MAC = %q", got.MAC)
	}
	if got.Duration != 24*time.Hour {
		t.Fatalf("duration = %s, want 24h", got.Duration)
	}
}

func TestPixDevAprovar_DevelopmentAtualizaStatusParaAprovado(t *testing.T) {
	app := fiber.New()
	appStore := &fakeStore{
		pix: map[string]store.PixTransaction{
			"ast_dev": {TXID: "ast_dev", Status: "pendente"},
		},
	}
	portal.Register(app, portal.Dependencies{Store: appStore, Env: "development"})

	req := httptest.NewRequest("POST", "/api/pix/dev/aprovar/ast_dev", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Fatalf("esperava status 200, got %d body=%s", resp.StatusCode, string(body))
	}
	if len(appStore.updatedPix) != 1 {
		t.Fatalf("updates len = %d, want 1", len(appStore.updatedPix))
	}
	if appStore.updatedPix[0].TXID != "ast_dev" || appStore.updatedPix[0].Status != "aprovado" {
		t.Fatalf("update incorreto: %+v", appStore.updatedPix[0])
	}
}

func TestPixStatus_DepoisDeAprovacaoDevRetornaAprovado(t *testing.T) {
	app := fiber.New()
	appStore := memory.NewStore()
	portal.Register(app, portal.Dependencies{Store: appStore, Env: "development"})

	createReq := httptest.NewRequest("POST", "/api/pix/gerar", strings.NewReader(`{
		"plano_id": 1,
		"mac": "AA:BB:CC:DD:EE:FF",
		"ip": "192.168.1.50",
		"nome": "Smoke"
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	createResp, err := app.Test(createReq, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer createResp.Body.Close()
	var created store.PixTransaction
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}

	approveReq := httptest.NewRequest("POST", "/api/pix/dev/aprovar/"+created.TXID, nil)
	approveResp, err := app.Test(approveReq, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer approveResp.Body.Close()
	if approveResp.StatusCode != 200 {
		body, _ := io.ReadAll(approveResp.Body)
		t.Fatalf("esperava aprovar PIX, got %d body=%s", approveResp.StatusCode, string(body))
	}
	if direct, ok, err := appStore.PixStatus(context.Background(), created.TXID); err != nil || !ok || direct.Status != "aprovado" {
		t.Fatalf("store direto apos aprovacao = (%+v, %v, %v), want aprovado", direct, ok, err)
	}

	statusReq := httptest.NewRequest("GET", "/api/pix/status/"+created.TXID, nil)
	statusResp, err := app.Test(statusReq, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer statusResp.Body.Close()
	body, _ := io.ReadAll(statusResp.Body)

	if statusResp.StatusCode != 200 {
		t.Fatalf("esperava status 200, got %d body=%s", statusResp.StatusCode, string(body))
	}
	if !strings.Contains(string(body), `"status":"aprovado"`) {
		t.Fatalf("esperava status aprovado, got %s", string(body))
	}
}

func TestGerarPix_UsaPaymentProviderConfigurado(t *testing.T) {
	app := fiber.New()
	appStore := memory.NewStore()
	provider := &fakePaymentProvider{
		createPix: payments.Pix{
			PixCopiaCola: "pix-real",
			QRCodeBase64: "qr-real",
		},
	}
	portal.Register(app, portal.Dependencies{
		Store:           appStore,
		PaymentProvider: provider,
	})

	req := httptest.NewRequest("POST", "/api/pix/gerar", strings.NewReader(`{
		"plano_id": 1,
		"mac": "AA:BB:CC:DD:EE:FF",
		"ip": "192.168.1.50",
		"nome": "Cliente"
	}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 201 {
		t.Fatalf("esperava status 201, got %d body=%s", resp.StatusCode, string(body))
	}
	if len(provider.createdInputs) != 1 {
		t.Fatalf("created inputs len = %d, want 1", len(provider.createdInputs))
	}
	if provider.createdInputs[0].TXID == "" {
		t.Fatal("provider recebeu TXID vazio")
	}
	if provider.createdInputs[0].Valor != "5.00" {
		t.Fatalf("provider Valor = %q, want 5.00", provider.createdInputs[0].Valor)
	}
	if provider.createdInputs[0].Descricao != "Astrolink Wi-Fi - Acesso 1 Hora" {
		t.Fatalf("provider Descricao = %q", provider.createdInputs[0].Descricao)
	}
	if !strings.Contains(string(body), `"pix_copia_cola":"pix-real"`) {
		t.Fatalf("resposta nao contem PIX do provider: %s", string(body))
	}
}

func TestPixDevAprovar_ProductionBloqueia(t *testing.T) {
	app := fiber.New()
	appStore := &fakeStore{}
	portal.Register(app, portal.Dependencies{Store: appStore, Env: "production"})

	req := httptest.NewRequest("POST", "/api/pix/dev/aprovar/ast_dev", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Fatalf("esperava status 404, got %d", resp.StatusCode)
	}
	if len(appStore.updatedPix) != 0 {
		t.Fatalf("updates len = %d, want 0", len(appStore.updatedPix))
	}
}

func TestMercadoPagoWebhook_SemSegredoNaoAprova(t *testing.T) {
	app := fiber.New()
	appStore := &fakeStore{}
	portal.Register(app, portal.Dependencies{
		Store:           appStore,
		Env:             "development",
		PaymentProvider: &fakePaymentProvider{status: payments.PixStatusApproved, txid: "ast_paid"},
	})

	req := httptest.NewRequest("POST", "/api/webhooks/mercadopago", strings.NewReader(`{"data":{"id":"123"}}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 202 {
		t.Fatalf("esperava status 202, got %d", resp.StatusCode)
	}
	if len(appStore.updatedPix) != 0 {
		t.Fatalf("updates len = %d, want 0", len(appStore.updatedPix))
	}
}

func TestMercadoPagoWebhook_AssinadoConsultaProviderEAprova(t *testing.T) {
	app := fiber.New()
	appStore := &fakeStore{
		pix: map[string]store.PixTransaction{
			"ast_paid": {TXID: "ast_paid", Status: "pendente"},
		},
	}
	secret := "mp-secret"
	paymentID := "123"
	requestID := "req-1"
	ts := "1716250000"
	portal.Register(app, portal.Dependencies{
		Store:                    appStore,
		Env:                      "production",
		MercadoPagoWebhookSecret: secret,
		PaymentProvider:          &fakePaymentProvider{status: payments.PixStatusApproved, txid: "ast_paid"},
	})

	req := httptest.NewRequest("POST", "/api/webhooks/mercadopago", strings.NewReader(`{"data":{"id":"123"}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-request-id", requestID)
	req.Header.Set("x-signature", "ts="+ts+",v1="+signPortalTestWebhook(secret, paymentID, requestID, ts))
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Fatalf("esperava status 200, got %d body=%s", resp.StatusCode, string(body))
	}
	if len(appStore.updatedPix) != 1 {
		t.Fatalf("updates len = %d, want 1", len(appStore.updatedPix))
	}
	if appStore.updatedPix[0].TXID != "ast_paid" || appStore.updatedPix[0].Status != "aprovado" {
		t.Fatalf("update incorreto: %+v", appStore.updatedPix[0])
	}
}

func TestMercadoPagoWebhook_AssinadoComIDNumericoGrande(t *testing.T) {
	app := fiber.New()
	appStore := &fakeStore{
		pix: map[string]store.PixTransaction{
			"ast_big": {TXID: "ast_big", Status: "pendente"},
		},
	}
	secret := "mp-secret"
	paymentID := "123456789012345678"
	requestID := "req-big"
	ts := "1716250000"
	portal.Register(app, portal.Dependencies{
		Store:                    appStore,
		Env:                      "production",
		MercadoPagoWebhookSecret: secret,
		PaymentProvider:          &fakePaymentProvider{status: payments.PixStatusApproved, txid: "ast_big"},
	})

	req := httptest.NewRequest("POST", "/api/webhooks/mercadopago", strings.NewReader(`{"data":{"id":123456789012345678}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-request-id", requestID)
	req.Header.Set("x-signature", "ts="+ts+",v1="+signPortalTestWebhook(secret, paymentID, requestID, ts))
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Fatalf("esperava status 200, got %d body=%s", resp.StatusCode, string(body))
	}
	if len(appStore.updatedPix) != 1 {
		t.Fatalf("updates len = %d, want 1", len(appStore.updatedPix))
	}
	if appStore.updatedPix[0].TXID != "ast_big" || appStore.updatedPix[0].Status != "aprovado" {
		t.Fatalf("update incorreto: %+v", appStore.updatedPix[0])
	}
}

type fakeStore struct {
	redeemResult store.RedeemVoucherResult
	redeemErr    error
	pix          map[string]store.PixTransaction
	updatedPix   []store.UpdatePixStatusInput
}

func (f *fakeStore) Settings(context.Context) (store.Settings, error) {
	return store.Settings{}, nil
}

func (f *fakeStore) PortalPlanos(context.Context) ([]planos.Plano, error) {
	return nil, nil
}

func (f *fakeStore) AdminPlanos(context.Context) ([]planos.Plano, error) {
	return nil, nil
}

func (f *fakeStore) AdminVouchers(context.Context) ([]store.AdminVoucher, error) {
	return nil, nil
}

func (f *fakeStore) GenerateVouchers(context.Context, store.GenerateVouchersInput) (store.GenerateVouchersResult, error) {
	return store.GenerateVouchersResult{}, nil
}

func (f *fakeStore) Usuarios(context.Context) ([]store.Usuario, error) {
	return nil, nil
}

func (f *fakeStore) SessaoStatus(context.Context, string) (store.Usuario, error) {
	return store.Usuario{Status: "walled_garden"}, nil
}

func (f *fakeStore) CreatePix(context.Context, store.CreatePixInput) (store.PixTransaction, error) {
	return store.PixTransaction{}, nil
}

func (f *fakeStore) PixStatus(_ context.Context, txid string) (store.PixTransaction, bool, error) {
	tx, ok := f.pix[txid]
	return tx, ok, nil
}

func (f *fakeStore) UpdatePixStatus(_ context.Context, input store.UpdatePixStatusInput) (store.PixTransaction, bool, error) {
	f.updatedPix = append(f.updatedPix, input)
	tx, ok := f.pix[input.TXID]
	if !ok {
		return store.PixTransaction{}, false, nil
	}
	tx.Status = input.Status
	f.pix[input.TXID] = tx
	return tx, true, nil
}

func (f *fakeStore) RedeemVoucher(context.Context, store.RedeemVoucherInput) (store.RedeemVoucherResult, error) {
	return f.redeemResult, f.redeemErr
}

func (f *fakeStore) Health(context.Context) store.Health {
	return store.Health{}
}

type fakePaymentProvider struct {
	status        payments.PixStatus
	txid          string
	createPix     payments.Pix
	createdInputs []payments.CreatePixInput
}

func (f *fakePaymentProvider) CreatePix(_ context.Context, input payments.CreatePixInput) (payments.Pix, error) {
	f.createdInputs = append(f.createdInputs, input)
	result := f.createPix
	if result.TXID == "" {
		result.TXID = input.TXID
	}
	if result.ExpiresAt.IsZero() {
		result.ExpiresAt = input.ExpiresAt
	}
	return result, nil
}

func (f fakePaymentProvider) PixStatus(context.Context, payments.StatusQuery) (payments.StatusResult, error) {
	return payments.StatusResult{TXID: f.txid, Status: f.status}, nil
}

func signPortalTestWebhook(secret, dataID, requestID, ts string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("id:" + dataID + ";request-id:" + requestID + ";ts:" + ts + ";"))
	return hex.EncodeToString(mac.Sum(nil))
}

type fakeGateway struct {
	authorizations []gateway.Authorization
	deauths        []string
	authErr        error
	deauthErr      error
}

func (f *fakeGateway) Authorize(_ context.Context, input gateway.Authorization) error {
	f.authorizations = append(f.authorizations, input)
	return f.authErr
}

func (f *fakeGateway) Deauthorize(_ context.Context, mac string) error {
	f.deauths = append(f.deauths, mac)
	return f.deauthErr
}

func (f *fakeGateway) Ping(context.Context) (time.Duration, error) {
	return 0, nil
}
