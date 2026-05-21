package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/astrolink/node/internal/domain/planos"
	"github.com/astrolink/node/internal/gateway"
	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

func TestHealthHandler_ConsultaGatewayOnline(t *testing.T) {
	app := fiber.New()
	app.Get("/health", healthHandler(Dependencies{
		Store:   adminRouterTestStore{health: store.Health{DatabaseStatus: "ok", DatabaseLatencyMS: 3}},
		Gateway: &adminRouterTestGateway{latency: 42 * time.Millisecond},
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/health", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var got struct {
		Checks struct {
			Roteadores struct {
				Total     int    `json:"total"`
				Online    int    `json:"online"`
				Offline   int    `json:"offline"`
				Status    string `json:"status"`
				LatencyMS int64  `json:"latencia_ms"`
			} `json:"roteadores"`
		} `json:"checks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Checks.Roteadores.Total != 1 || got.Checks.Roteadores.Online != 1 || got.Checks.Roteadores.Offline != 0 {
		t.Fatalf("roteadores = %+v", got.Checks.Roteadores)
	}
	if got.Checks.Roteadores.Status != "online" || got.Checks.Roteadores.LatencyMS != 42 {
		t.Fatalf("status/latencia = %+v", got.Checks.Roteadores)
	}
}

func TestHealthHandler_ReportaGatewayOffline(t *testing.T) {
	app := fiber.New()
	app.Get("/health", healthHandler(Dependencies{
		Store:   adminRouterTestStore{health: store.Health{DatabaseStatus: "ok"}},
		Gateway: &adminRouterTestGateway{pingErr: errors.New("ndsctl unavailable")},
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/health", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var got struct {
		Checks struct {
			Roteadores struct {
				Online  int    `json:"online"`
				Offline int    `json:"offline"`
				Status  string `json:"status"`
			} `json:"roteadores"`
		} `json:"checks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Checks.Roteadores.Status != "offline" || got.Checks.Roteadores.Online != 0 || got.Checks.Roteadores.Offline != 1 {
		t.Fatalf("roteadores = %+v", got.Checks.Roteadores)
	}
}

func TestRoteadoresHandler_ReportaNoopComoDev(t *testing.T) {
	app := fiber.New()
	app.Get("/admin/roteadores", roteadoresHandler(Dependencies{
		Store:   adminRouterTestStore{},
		Gateway: gateway.NoopController{},
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/admin/roteadores", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var got struct {
		Roteadores []struct {
			ID     int    `json:"id"`
			Nome   string `json:"nome"`
			Status string `json:"status"`
		} `json:"roteadores"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got.Roteadores) != 1 {
		t.Fatalf("roteadores = %+v", got.Roteadores)
	}
	if got.Roteadores[0].ID != 1 || got.Roteadores[0].Status != "dev/disabled" {
		t.Fatalf("roteador = %+v", got.Roteadores[0])
	}
}

func TestRoteadorDiagnosticoHandler_UsaGatewayRealQuandoDisponivel(t *testing.T) {
	diagnostic := gateway.RouterDiagnostic{
		Online:      true,
		ClientCount: 1,
		OpenNDS:     gateway.OpenNDSStatus{Online: true, Version: "10.2.0", ClientCount: 1},
		Clients:     []gateway.ClientSummary{{MAC: "AA:BB:CC:DD:EE:FF", IP: "192.168.1.23"}},
	}
	app := fiber.New()
	app.Get("/admin/roteadores/:id/diagnostico", roteadorDiagnosticoHandler(Dependencies{
		Store:   adminRouterTestStore{},
		Gateway: &adminRouterTestGateway{diagnostic: diagnostic},
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/admin/roteadores/1/diagnostico", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var got struct {
		Status      string                   `json:"status"`
		Diagnostico gateway.RouterDiagnostic `json:"diagnostico"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Status != "online" || got.Diagnostico.OpenNDS.Version != "10.2.0" || got.Diagnostico.ClientCount != 1 {
		t.Fatalf("diagnostico = %+v", got)
	}
}

func TestRoteadorDiagnosticoHandler_RetornaDevSemCapacidade(t *testing.T) {
	app := fiber.New()
	app.Get("/admin/roteadores/:id/diagnostico", roteadorDiagnosticoHandler(Dependencies{
		Store:   adminRouterTestStore{},
		Gateway: &adminRouterPingOnlyGateway{},
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/admin/roteadores/1/diagnostico", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var got struct {
		Status      string                   `json:"status"`
		Diagnostico gateway.RouterDiagnostic `json:"diagnostico"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Status != "dev/disabled" || got.Diagnostico.Online || len(got.Diagnostico.RecentLogs) != 0 {
		t.Fatalf("diagnostico dev = %+v", got)
	}
}

type adminRouterTestStore struct {
	health store.Health
}

func (s adminRouterTestStore) Settings(context.Context) (store.Settings, error) {
	return store.Settings{}, nil
}
func (s adminRouterTestStore) PortalPlanos(context.Context) ([]planos.Plano, error) { return nil, nil }
func (s adminRouterTestStore) AdminPlanos(context.Context) ([]planos.Plano, error)  { return nil, nil }
func (s adminRouterTestStore) Usuarios(context.Context) ([]store.Usuario, error)    { return nil, nil }
func (s adminRouterTestStore) SessaoStatus(context.Context, string) (store.Usuario, error) {
	return store.Usuario{}, nil
}
func (s adminRouterTestStore) CreatePix(context.Context, store.CreatePixInput) (store.PixTransaction, error) {
	return store.PixTransaction{}, nil
}
func (s adminRouterTestStore) PixStatus(context.Context, string) (store.PixTransaction, bool, error) {
	return store.PixTransaction{}, false, nil
}
func (s adminRouterTestStore) RedeemVoucher(context.Context, store.RedeemVoucherInput) (store.RedeemVoucherResult, error) {
	return store.RedeemVoucherResult{}, nil
}
func (s adminRouterTestStore) AdminVouchers(context.Context) ([]store.AdminVoucher, error) {
	return nil, nil
}
func (s adminRouterTestStore) GenerateVouchers(context.Context, store.GenerateVouchersInput) (store.GenerateVouchersResult, error) {
	return store.GenerateVouchersResult{}, nil
}
func (s adminRouterTestStore) Health(context.Context) store.Health { return s.health }

type adminRouterTestGateway struct {
	latency    time.Duration
	pingErr    error
	diagnostic gateway.RouterDiagnostic
}

func (g *adminRouterTestGateway) Authorize(context.Context, gateway.Authorization) error { return nil }
func (g *adminRouterTestGateway) Deauthorize(context.Context, string) error              { return nil }
func (g *adminRouterTestGateway) Ping(context.Context) (time.Duration, error) {
	return g.latency, g.pingErr
}
func (g *adminRouterTestGateway) Diagnostic(context.Context) (gateway.RouterDiagnostic, error) {
	return g.diagnostic, nil
}

type adminRouterPingOnlyGateway struct{}

func (g *adminRouterPingOnlyGateway) Authorize(context.Context, gateway.Authorization) error {
	return nil
}
func (g *adminRouterPingOnlyGateway) Deauthorize(context.Context, string) error   { return nil }
func (g *adminRouterPingOnlyGateway) Ping(context.Context) (time.Duration, error) { return 0, nil }
