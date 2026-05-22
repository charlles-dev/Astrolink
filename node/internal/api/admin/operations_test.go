package admin

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/astrolink/node/internal/config"
	"github.com/astrolink/node/internal/domain/planos"
	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

func TestLogsHandler_RetornaFallbackOperacionalFiltrado(t *testing.T) {
	app := fiber.New()
	deps := Dependencies{Config: operationsTestConfig(), Store: operationsStore{}}
	app.Get("/admin/logs", logsHandler(deps))

	req := httptest.NewRequest("GET", "/admin/logs?nivel=info&tipo=sistema&texto=dev", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	var got struct {
		Total int              `json:"total"`
		Logs  []OperationalLog `json:"logs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Total != 1 || len(got.Logs) != 1 {
		t.Fatalf("logs = %+v", got)
	}
	if got.Logs[0].Nivel != "info" || got.Logs[0].Tipo != "sistema" || !strings.Contains(got.Logs[0].Mensagem, "dev") {
		t.Fatalf("log inesperado: %+v", got.Logs[0])
	}
}

func TestLogsCSVHandler_RetornaCSVFiltrado(t *testing.T) {
	app := fiber.New()
	deps := Dependencies{Config: operationsTestConfig(), Store: operationsStore{}}
	app.Get("/admin/logs/export.csv", exportLogsCSVHandler(deps))

	req := httptest.NewRequest("GET", "/admin/logs/export.csv?tipo=backup", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "text/csv") {
		t.Fatalf("content-type = %q", contentType)
	}
	if !strings.Contains(string(body), "timestamp,nivel,tipo,mensagem,detalhes") {
		t.Fatalf("csv sem cabecalho esperado: %s", string(body))
	}
	if !strings.Contains(string(body), "backup manual requer Postgres") {
		t.Fatalf("csv sem log esperado: %s", string(body))
	}
}

func TestBackupHandler_RetornaErroControladoQuandoStoreNaoSuportaBackup(t *testing.T) {
	app := fiber.New()
	deps := Dependencies{Config: operationsTestConfig(), Store: operationsStore{}}
	app.Post("/admin/backup", backupHandler(deps))

	req := httptest.NewRequest("POST", "/admin/backup", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusNotImplemented {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	var got struct {
		Erro     string `json:"erro"`
		Mensagem string `json:"mensagem"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Erro == "" || !strings.Contains(got.Mensagem, "Postgres") || strings.Contains(strings.ToLower(got.Mensagem), "restore") {
		t.Fatalf("erro inesperado: %+v", got)
	}
}

func TestRestoreBackupHandler_ExigeConfirmacaoExata(t *testing.T) {
	app := fiber.New()
	deps := Dependencies{Config: operationsTestConfig(), Store: operationsStore{}}
	app.Post("/admin/backup/restaurar", restoreBackupHandler(deps))

	req := httptest.NewRequest("POST", "/admin/backup/restaurar", strings.NewReader(`{"arquivo":"backup.sql","confirmacao":"restaurar"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	var got struct {
		Erro     string `json:"erro"`
		Mensagem string `json:"mensagem"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Erro == "" || !strings.Contains(got.Mensagem, "RESTAURAR") {
		t.Fatalf("erro inesperado: %+v", got)
	}
}

func TestRestoreBackupHandler_RetornaIndisponivelMesmoComConfirmacao(t *testing.T) {
	app := fiber.New()
	deps := Dependencies{Config: operationsTestConfig(), Store: operationsStore{}}
	app.Post("/admin/backup/restaurar", restoreBackupHandler(deps))

	req := httptest.NewRequest("POST", "/admin/backup/restaurar", strings.NewReader(`{"arquivo":"backup.sql","confirmacao":"RESTAURAR"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusNotImplemented {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	var got struct {
		Erro     string `json:"erro"`
		Mensagem string `json:"mensagem"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Erro != "restore_indisponivel" || !strings.Contains(got.Mensagem, "Postgres") || !strings.Contains(got.Mensagem, "manual") {
		t.Fatalf("erro inesperado: %+v", got)
	}
}

type operationsStore struct{}

func (operationsStore) Settings(context.Context) (store.Settings, error) {
	return store.Settings{}, nil
}
func (operationsStore) PortalPlanos(context.Context) ([]planos.Plano, error) { return nil, nil }
func (operationsStore) AdminPlanos(context.Context) ([]planos.Plano, error)  { return nil, nil }
func (operationsStore) Usuarios(context.Context) ([]store.Usuario, error)    { return nil, nil }
func (operationsStore) SessaoStatus(context.Context, string) (store.Usuario, error) {
	return store.Usuario{}, nil
}
func (operationsStore) CreatePix(context.Context, store.CreatePixInput) (store.PixTransaction, error) {
	return store.PixTransaction{}, nil
}
func (operationsStore) PixStatus(context.Context, string) (store.PixTransaction, bool, error) {
	return store.PixTransaction{}, false, nil
}
func (operationsStore) UpdatePixStatus(context.Context, store.UpdatePixStatusInput) (store.PixTransaction, bool, error) {
	return store.PixTransaction{}, false, nil
}
func (operationsStore) RedeemVoucher(context.Context, store.RedeemVoucherInput) (store.RedeemVoucherResult, error) {
	return store.RedeemVoucherResult{}, nil
}
func (operationsStore) AdminVouchers(context.Context) ([]store.AdminVoucher, error) { return nil, nil }
func (operationsStore) GenerateVouchers(context.Context, store.GenerateVouchersInput) (store.GenerateVouchersResult, error) {
	return store.GenerateVouchersResult{}, nil
}
func (operationsStore) Health(context.Context) store.Health {
	return store.Health{DatabaseStatus: "memory"}
}

func operationsTestConfig() config.Config {
	return config.Config{
		AdminUser:     "admin",
		AdminPassword: "admin123",
		JWTSecret:     "test-jwt-secret-com-mais-de-32-bytes",
	}
}
