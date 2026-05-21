package portal_test

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/astrolink/node/internal/api/portal"
	"github.com/astrolink/node/internal/domain/planos"
	"github.com/astrolink/node/internal/gateway"
	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

func TestResgatarVoucher_CodigoInexistente_Retorna404(t *testing.T) {
	app := fiber.New()
	portal.Register(app, portal.Dependencies{Store: fakeStore{redeemErr: store.ErrVoucherNotFound}})

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
	appStore := fakeStore{
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

type fakeStore struct {
	redeemResult store.RedeemVoucherResult
	redeemErr    error
}

func (f fakeStore) Settings(context.Context) (store.Settings, error) {
	return store.Settings{}, nil
}

func (f fakeStore) PortalPlanos(context.Context) ([]planos.Plano, error) {
	return nil, nil
}

func (f fakeStore) AdminPlanos(context.Context) ([]planos.Plano, error) {
	return nil, nil
}

func (f fakeStore) Usuarios(context.Context) ([]store.Usuario, error) {
	return nil, nil
}

func (f fakeStore) SessaoStatus(context.Context, string) (store.Usuario, error) {
	return store.Usuario{Status: "walled_garden"}, nil
}

func (f fakeStore) CreatePix(context.Context, store.CreatePixInput) (store.PixTransaction, error) {
	return store.PixTransaction{}, nil
}

func (f fakeStore) PixStatus(context.Context, string) (store.PixTransaction, bool, error) {
	return store.PixTransaction{}, false, nil
}

func (f fakeStore) RedeemVoucher(context.Context, store.RedeemVoucherInput) (store.RedeemVoucherResult, error) {
	return f.redeemResult, f.redeemErr
}

func (f fakeStore) Health(context.Context) store.Health {
	return store.Health{}
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
