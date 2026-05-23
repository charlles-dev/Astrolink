package admin

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/astrolink/node/internal/config"
	"github.com/astrolink/node/internal/infra/memory"
	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

func TestVouchersHandler_FiltraPorStatusPlanoCodigoELote(t *testing.T) {
	app, repo := newVoucherTestApp()
	firstLot := mustGenerateVouchers(t, repo, store.GenerateVouchersInput{PlanoID: 2, Quantidade: 2, Prefixo: "VIP"})
	mustGenerateVouchers(t, repo, store.GenerateVouchersInput{PlanoID: 1, Quantidade: 1, Prefixo: "STAFF"})
	mustDeactivateVoucher(t, repo, firstLot.Vouchers[0].ID)

	req := httptest.NewRequest("GET", "/admin/vouchers?status=ativo&plano_id=2&codigo=vip&lote_id="+strconv.Itoa(firstLot.LoteID), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	var got struct {
		Total    int                  `json:"total"`
		Vouchers []store.AdminVoucher `json:"vouchers"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.Total != 1 || len(got.Vouchers) != 1 {
		t.Fatalf("total = %d len = %d body=%s", got.Total, len(got.Vouchers), string(body))
	}
	if got.Vouchers[0].Plano.ID != 2 || got.Vouchers[0].LoteID == nil || *got.Vouchers[0].LoteID != firstLot.LoteID {
		t.Fatalf("voucher filtrado inesperado: %+v", got.Vouchers[0])
	}
	if !strings.Contains(strings.ToLower(got.Vouchers[0].Codigo), "vip") || !got.Vouchers[0].Ativo {
		t.Fatalf("codigo/status inesperado: %+v", got.Vouchers[0])
	}
}

func TestDesativarVoucherHandler_RetornaVoucherInativo(t *testing.T) {
	app, repo := newVoucherTestApp()
	generated := mustGenerateVouchers(t, repo, store.GenerateVouchersInput{PlanoID: 2, Quantidade: 1, Prefixo: "OFF"})

	req := httptest.NewRequest("PATCH", "/admin/vouchers/"+strconv.Itoa(generated.Vouchers[0].ID)+"/desativar", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	var got struct {
		Voucher store.AdminVoucher `json:"voucher"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.Voucher.ID != generated.Vouchers[0].ID || got.Voucher.Ativo {
		t.Fatalf("voucher = %+v", got.Voucher)
	}
}

func TestExportVouchersCSVHandler_UsaFiltrosECabecalho(t *testing.T) {
	app, repo := newVoucherTestApp()
	generated := mustGenerateVouchers(t, repo, store.GenerateVouchersInput{PlanoID: 2, Quantidade: 1, Prefixo: "CSV"})
	mustGenerateVouchers(t, repo, store.GenerateVouchersInput{PlanoID: 1, Quantidade: 1, Prefixo: "OTHER"})

	req := httptest.NewRequest("GET", "/admin/vouchers/export.csv?status=ativo&plano_id=2&codigo=csv&lote_id="+strconv.Itoa(generated.LoteID), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "text/csv") {
		t.Fatalf("content-type = %q", contentType)
	}
	records, err := csv.NewReader(strings.NewReader(string(body))).ReadAll()
	if err != nil {
		t.Fatalf("csv invalido: %v body=%s", err, string(body))
	}
	wantHeader := []string{"codigo", "plano", "tipo", "usos_atuais", "usos_maximos", "ativo", "validade_em", "prefixo", "lote_id", "created_at"}
	if len(records) != 2 {
		t.Fatalf("records = %d want 2 body=%s", len(records), string(body))
	}
	for i, want := range wantHeader {
		if records[0][i] != want {
			t.Fatalf("header[%d] = %q want %q", i, records[0][i], want)
		}
	}
	if records[1][0] != generated.Vouchers[0].Codigo || records[1][1] != "Acesso 24 Horas" || records[1][7] != "CSV" || records[1][8] != strconv.Itoa(generated.LoteID) {
		t.Fatalf("linha inesperada: %#v", records[1])
	}
}

func newVoucherTestApp() (*fiber.App, *memory.Store) {
	repo := memory.NewStore()
	app := fiber.New()
	deps := Dependencies{Config: config.Config{}, Store: repo}
	app.Get("/admin/vouchers", vouchersHandler(deps))
	app.Patch("/admin/vouchers/:id/desativar", desativarVoucherHandler(deps))
	app.Get("/admin/vouchers/export.csv", exportVouchersCSVHandler(deps))
	return app, repo
}

func mustGenerateVouchers(t *testing.T, repo *memory.Store, input store.GenerateVouchersInput) store.GenerateVouchersResult {
	t.Helper()
	result, err := repo.GenerateVouchers(t.Context(), input)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func mustDeactivateVoucher(t *testing.T, repo *memory.Store, id int) store.AdminVoucher {
	t.Helper()
	voucher, err := repo.DeactivateVoucher(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	return voucher
}
