package admin_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/astrolink/node/internal/api/admin"
	"github.com/astrolink/node/internal/config"
	"github.com/astrolink/node/internal/domain/planos"
	"github.com/astrolink/node/internal/gateway"
	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

func TestDesconectarUsuario_ChamaGatewayDeauthorize(t *testing.T) {
	app := fiber.New()
	router := &fakeGateway{}
	admin.Register(app, admin.Dependencies{
		Config:  config.Config{AdminUser: "admin", AdminPassword: "admin123"},
		Store:   &fakeStore{},
		Gateway: router,
	})

	req := httptest.NewRequest("POST", "/admin/usuarios/AA:BB:CC:DD:EE:FF/desconectar", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	if !strings.Contains(string(body), `"sucesso":true`) {
		t.Fatalf("resposta inesperada: %s", string(body))
	}
	if len(router.deauths) != 1 || router.deauths[0] != "AA:BB:CC:DD:EE:FF" {
		t.Fatalf("deauths = %#v", router.deauths)
	}
}

func TestListarVouchers_RetornaVouchersDoStore(t *testing.T) {
	app := fiber.New()
	admin.Register(app, admin.Dependencies{
		Config: config.Config{AdminUser: "admin", AdminPassword: "admin123"},
		Store: &fakeStore{
			vouchers: []store.AdminVoucher{
				{
					ID:     1,
					Codigo: "VIPA-1234",
					Plano:  store.PlanoResumo{ID: 2, Nome: "Acesso 24 Horas"},
					Tipo:   "single_use",
					Ativo:  true,
				},
			},
		},
		Gateway: &fakeGateway{},
	})

	req := httptest.NewRequest("GET", "/admin/vouchers", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	if !strings.Contains(string(body), `"codigo":"VIPA-1234"`) {
		t.Fatalf("resposta inesperada: %s", string(body))
	}
}

func TestGerarVouchers_RetornaCodigosCriados(t *testing.T) {
	app := fiber.New()
	repo := &fakeStore{
		generated: store.GenerateVouchersResult{
			LoteID:     7,
			Quantidade: 2,
			Vouchers: []store.AdminVoucher{
				{ID: 3, Codigo: "VIPA-1111", Plano: store.PlanoResumo{ID: 2, Nome: "Acesso 24 Horas"}, Tipo: "single_use", Ativo: true},
				{ID: 4, Codigo: "VIPA-2222", Plano: store.PlanoResumo{ID: 2, Nome: "Acesso 24 Horas"}, Tipo: "single_use", Ativo: true},
			},
		},
	}
	admin.Register(app, admin.Dependencies{
		Config:  config.Config{AdminUser: "admin", AdminPassword: "admin123"},
		Store:   repo,
		Gateway: &fakeGateway{},
	})

	body := strings.NewReader(`{"plano_id":2,"quantidade":2,"prefixo":"VIPA"}`)
	req := httptest.NewRequest("POST", "/admin/vouchers/gerar", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(payload))
	}
	if repo.generateInput.PlanoID != 2 || repo.generateInput.Quantidade != 2 || repo.generateInput.Prefixo != "VIPA" {
		t.Fatalf("input recebido = %+v", repo.generateInput)
	}
	var got store.GenerateVouchersResult
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.LoteID != 7 || got.Quantidade != 2 || len(got.Vouchers) != 2 {
		t.Fatalf("resposta inesperada: %+v", got)
	}
}

type fakeStore struct {
	vouchers      []store.AdminVoucher
	generated     store.GenerateVouchersResult
	generateInput store.GenerateVouchersInput
}

func (fakeStore) Settings(context.Context) (store.Settings, error)     { return store.Settings{}, nil }
func (fakeStore) PortalPlanos(context.Context) ([]planos.Plano, error) { return nil, nil }
func (fakeStore) AdminPlanos(context.Context) ([]planos.Plano, error)  { return nil, nil }
func (fakeStore) Usuarios(context.Context) ([]store.Usuario, error)    { return nil, nil }
func (fakeStore) SessaoStatus(context.Context, string) (store.Usuario, error) {
	return store.Usuario{}, nil
}
func (fakeStore) CreatePix(context.Context, store.CreatePixInput) (store.PixTransaction, error) {
	return store.PixTransaction{}, nil
}
func (fakeStore) PixStatus(context.Context, string) (store.PixTransaction, bool, error) {
	return store.PixTransaction{}, false, nil
}
func (fakeStore) RedeemVoucher(context.Context, store.RedeemVoucherInput) (store.RedeemVoucherResult, error) {
	return store.RedeemVoucherResult{}, nil
}
func (fakeStore) Health(context.Context) store.Health { return store.Health{} }

func (f fakeStore) AdminVouchers(context.Context) ([]store.AdminVoucher, error) {
	return f.vouchers, nil
}

func (f *fakeStore) GenerateVouchers(_ context.Context, input store.GenerateVouchersInput) (store.GenerateVouchersResult, error) {
	f.generateInput = input
	return f.generated, nil
}

type fakeGateway struct {
	authorizations []gateway.Authorization
	deauths        []string
}

func (f *fakeGateway) Authorize(_ context.Context, input gateway.Authorization) error {
	f.authorizations = append(f.authorizations, input)
	return nil
}

func (f *fakeGateway) Deauthorize(_ context.Context, mac string) error {
	f.deauths = append(f.deauths, mac)
	return nil
}

func (*fakeGateway) Ping(context.Context) (time.Duration, error) {
	return 0, nil
}
