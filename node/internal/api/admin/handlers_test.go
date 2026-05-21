package admin_test

import (
	"context"
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
		Store:   fakeStore{},
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

type fakeStore struct{}

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
