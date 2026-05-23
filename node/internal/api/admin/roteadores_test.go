package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/astrolink/node/internal/domain/planos"
	"github.com/astrolink/node/internal/gateway"
	"github.com/astrolink/node/internal/infra/memory"
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

func TestRedeRoteadoresHandlers_GerenciamCadastroLocal(t *testing.T) {
	app := fiber.New()
	repo := memory.NewStore()
	deps := Dependencies{Store: repo, Gateway: gateway.NoopController{}}
	app.Get("/admin/rede/roteadores", roteadoresHandler(deps))
	app.Post("/admin/rede/roteadores", criarRoteadorHandler(deps))
	app.Put("/admin/rede/roteadores/:id", atualizarRoteadorHandler(deps))
	app.Delete("/admin/rede/roteadores/:id", removerRoteadorHandler(deps))

	createReq := httptest.NewRequest("POST", "/admin/rede/roteadores", strings.NewReader(`{
		"nome":"Roteador Patio",
		"ip":"192.168.1.2",
		"porta_ssh":22,
		"usuario_ssh":"root",
		"chave_ssh_path":"",
		"ativo":true
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(createReq, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status create = %d", resp.StatusCode)
	}

	var created struct {
		Roteador store.AdminRoteador `json:"roteador"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.Roteador.Nome != "Roteador Patio" || created.Roteador.IP != "192.168.1.2" {
		t.Fatalf("roteador criado = %+v", created.Roteador)
	}

	updateReq := httptest.NewRequest("PUT", "/admin/rede/roteadores/2", strings.NewReader(`{
		"nome":"Roteador Patio 2",
		"ip":"192.168.1.22",
		"porta_ssh":2222,
		"usuario_ssh":"admin",
		"chave_ssh_path":"/etc/astrolink/key",
		"ativo":true
	}`))
	updateReq.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(updateReq, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status update = %d", resp.StatusCode)
	}

	resp, err = app.Test(httptest.NewRequest("DELETE", "/admin/rede/roteadores/2", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status delete = %d", resp.StatusCode)
	}
}

func TestRedeListasHandlers_GerenciamBlacklistEWalledGarden(t *testing.T) {
	app := fiber.New()
	repo := memory.NewStore()
	deps := Dependencies{Store: repo, Gateway: gateway.NoopController{}}
	app.Get("/admin/rede/blacklist", blacklistHandler(deps))
	app.Post("/admin/rede/blacklist", adicionarBlacklistHandler(deps, gateway.NoopController{}))
	app.Delete("/admin/rede/blacklist/:mac", removerBlacklistHandler(deps))
	app.Get("/admin/rede/walled-garden", walledGardenHandler(deps))
	app.Post("/admin/rede/walled-garden", adicionarWalledGardenHandler(deps))
	app.Delete("/admin/rede/walled-garden/:id", removerWalledGardenHandler(deps))

	blacklistReq := httptest.NewRequest("POST", "/admin/rede/blacklist", strings.NewReader(`{
		"mac":"AA:BB:CC:DD:EE:FF",
		"motivo":"teste"
	}`))
	blacklistReq.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(blacklistReq, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status blacklist = %d", resp.StatusCode)
	}

	var blacklist struct {
		Entrada store.AdminBlacklistEntry `json:"entrada"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&blacklist); err != nil {
		t.Fatal(err)
	}
	if blacklist.Entrada.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Fatalf("entrada blacklist = %+v", blacklist.Entrada)
	}

	gardenReq := httptest.NewRequest("POST", "/admin/rede/walled-garden", strings.NewReader(`{
		"host":"status.astrolink.local",
		"descricao":"Status local",
		"tipo":"dominio"
	}`))
	gardenReq.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(gardenReq, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status walled garden = %d", resp.StatusCode)
	}

	var garden struct {
		Entrada store.AdminWalledGardenEntry `json:"entrada"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&garden); err != nil {
		t.Fatal(err)
	}
	if garden.Entrada.Host != "status.astrolink.local" || garden.Entrada.Sistema {
		t.Fatalf("entrada garden = %+v", garden.Entrada)
	}

	resp, err = app.Test(httptest.NewRequest("DELETE", "/admin/rede/blacklist/AA:BB:CC:DD:EE:FF", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status delete blacklist = %d", resp.StatusCode)
	}

	resp, err = app.Test(httptest.NewRequest("DELETE", "/admin/rede/walled-garden/"+strconv.Itoa(garden.Entrada.ID), nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status delete garden = %d", resp.StatusCode)
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
func (s adminRouterTestStore) UpdatePixStatus(context.Context, store.UpdatePixStatusInput) (store.PixTransaction, bool, error) {
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
