package admin

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/astrolink/node/internal/config"
	"github.com/astrolink/node/internal/infra/memory"
	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

func TestEventsHandler_EmiteSnapshotSSEUmaVez(t *testing.T) {
	app, repo := newEventsTestApp()
	mustCreateEventPix(t, repo, store.CreatePixInput{PlanoID: 2, MAC: "aa:bb:cc:dd:ee:11"})
	approved := mustCreateEventPix(t, repo, store.CreatePixInput{PlanoID: 1, MAC: "aa:bb:cc:dd:ee:12"})
	mustUpdateEventPixStatus(t, repo, approved.TXID, "aprovado")
	mustRedeemEventVoucher(t, repo, store.RedeemVoucherInput{
		Codigo: "TEST-1234",
		MAC:    "aa:bb:cc:dd:ee:13",
		IP:     "192.168.0.13",
	})

	req := httptest.NewRequest("GET", "/admin/eventos?once=1", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "text/event-stream") {
		t.Fatalf("content-type = %q", contentType)
	}
	if cacheControl := resp.Header.Get("Cache-Control"); cacheControl != "no-cache" {
		t.Fatalf("cache-control = %q", cacheControl)
	}
	if connection := resp.Header.Get("Connection"); connection != "keep-alive" {
		t.Fatalf("connection = %q", connection)
	}
	if !strings.Contains(string(body), "event: snapshot\n") {
		t.Fatalf("evento snapshot ausente: %s", string(body))
	}

	var snapshot eventSnapshot
	if err := json.Unmarshal([]byte(sseDataLine(t, string(body))), &snapshot); err != nil {
		t.Fatalf("json invalido: %v body=%s", err, string(body))
	}
	if _, err := time.Parse(time.RFC3339, snapshot.Timestamp); err != nil {
		t.Fatalf("timestamp nao RFC3339: %q", snapshot.Timestamp)
	}
	if snapshot.Database != "memory" {
		t.Fatalf("database = %q", snapshot.Database)
	}
	if snapshot.UsuariosTotal != 1 || snapshot.UsuariosAtivos != 1 {
		t.Fatalf("usuarios inesperados: total=%d ativos=%d", snapshot.UsuariosTotal, snapshot.UsuariosAtivos)
	}
	if snapshot.VouchersTotal < 2 || snapshot.VouchersAtivos < 2 {
		t.Fatalf("vouchers inesperados: total=%d ativos=%d", snapshot.VouchersTotal, snapshot.VouchersAtivos)
	}
	if snapshot.PagamentosTotal != 2 || snapshot.PagamentosPendentes != 1 || snapshot.PagamentosAprovados != 1 {
		t.Fatalf("pagamentos inesperados: total=%d pendentes=%d aprovados=%d", snapshot.PagamentosTotal, snapshot.PagamentosPendentes, snapshot.PagamentosAprovados)
	}
	if snapshot.LogsTotal == 0 {
		t.Fatalf("logs_total = %d", snapshot.LogsTotal)
	}
}

func newEventsTestApp() (*fiber.App, *memory.Store) {
	repo := memory.NewStore()
	app := fiber.New()
	deps := Dependencies{Config: config.Config{}, Store: repo}
	app.Get("/admin/eventos", eventsHandler(deps))
	return app, repo
}

func mustCreateEventPix(t *testing.T, repo *memory.Store, input store.CreatePixInput) store.PixTransaction {
	t.Helper()
	tx, err := repo.CreatePix(t.Context(), input)
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

func mustUpdateEventPixStatus(t *testing.T, repo *memory.Store, txid string, status string) store.PixTransaction {
	t.Helper()
	tx, ok, err := repo.UpdatePixStatus(t.Context(), store.UpdatePixStatusInput{TXID: txid, Status: status})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("pix %q nao encontrado", txid)
	}
	return tx
}

func mustRedeemEventVoucher(t *testing.T, repo *memory.Store, input store.RedeemVoucherInput) store.RedeemVoucherResult {
	t.Helper()
	result, err := repo.RedeemVoucher(t.Context(), input)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func sseDataLine(t *testing.T, body string) string {
	t.Helper()
	for _, line := range strings.Split(body, "\n") {
		if data, ok := strings.CutPrefix(line, "data: "); ok {
			return data
		}
	}
	t.Fatalf("linha data ausente: %s", body)
	return ""
}
