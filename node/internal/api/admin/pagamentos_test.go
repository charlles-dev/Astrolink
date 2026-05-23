package admin

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/astrolink/node/internal/config"
	"github.com/astrolink/node/internal/infra/memory"
	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

func TestPagamentosHandler_FiltraStatusPeriodoERetornaTotais(t *testing.T) {
	app, repo := newPagamentosTestApp()
	mustCreatePix(t, repo, store.CreatePixInput{PlanoID: 2, MAC: "aa:bb:cc:dd:ee:01"})
	mustCreatePix(t, repo, store.CreatePixInput{PlanoID: 1, MAC: "aa:bb:cc:dd:ee:02"})

	req := httptest.NewRequest("GET", "/admin/pagamentos?status=pendente&inicio=2000-01-01&fim=2999-12-31", nil)
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
		Total      int                    `json:"total"`
		Totais     store.AdminPixTotals   `json:"totais"`
		Pagamentos []store.AdminPagamento `json:"pagamentos"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.Total != 2 || len(got.Pagamentos) != 2 {
		t.Fatalf("total = %d len = %d body=%s", got.Total, len(got.Pagamentos), string(body))
	}
	if got.Totais.Pendente != 2 || got.Totais.ValorTotal != "20.00" {
		t.Fatalf("totais inesperados: %+v", got.Totais)
	}
	payment := got.Pagamentos[0]
	if payment.TXID == "" || payment.Status != "pendente" || payment.Valor == "" || payment.MAC == "" {
		t.Fatalf("pagamento incompleto: %+v", payment)
	}
	if payment.Plano.ID == 0 || payment.Plano.Nome == "" || payment.PlanoID == 0 {
		t.Fatalf("plano nao mapeado: %+v", payment)
	}
	if payment.CreatedAt.IsZero() || payment.ExpiraEm.IsZero() {
		t.Fatalf("datas nao mapeadas: %+v", payment)
	}
}

func TestExportPagamentosCSVHandler_UsaFiltrosECabecalho(t *testing.T) {
	app, repo := newPagamentosTestApp()
	payment := mustCreatePix(t, repo, store.CreatePixInput{PlanoID: 2, MAC: "aa:bb:cc:dd:ee:03"})

	req := httptest.NewRequest("GET", "/admin/pagamentos/export.csv?status=todos", nil)
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
	wantHeader := []string{"txid", "status", "valor", "descricao", "mac", "plano_id", "plano", "created_at", "expira_em"}
	if len(records) != 2 {
		t.Fatalf("records = %d want 2 body=%s", len(records), string(body))
	}
	for i, want := range wantHeader {
		if records[0][i] != want {
			t.Fatalf("header[%d] = %q want %q", i, records[0][i], want)
		}
	}
	if records[1][0] != payment.TXID || records[1][1] != "pendente" || records[1][6] != "Acesso 24 Horas" {
		t.Fatalf("linha inesperada: %#v", records[1])
	}
}

func TestPagamentosRelatorioHandler_AceitaParametrosDaDoc(t *testing.T) {
	app, repo := newPagamentosTestApp()
	mustCreatePix(t, repo, store.CreatePixInput{PlanoID: 2, MAC: "aa:bb:cc:dd:ee:04"})

	req := httptest.NewRequest("GET", "/admin/pagamentos/relatorio?de=2000-01-01&ate=2999-12-31&formato=json", nil)
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
		Total      int                  `json:"total"`
		Totais     store.AdminPixTotals `json:"totais"`
		Pagamentos []struct {
			MAC string `json:"mac"`
		} `json:"pagamentos"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.Total != 1 || got.Totais.Pendente != 1 || got.Pagamentos[0].MAC != "AA:BB:CC:DD:EE:04" {
		t.Fatalf("relatorio inesperado: %+v body=%s", got, string(body))
	}
}

func TestPagamentosRelatorioHandler_GeraPDFLocal(t *testing.T) {
	app, repo := newPagamentosTestApp()
	mustCreatePix(t, repo, store.CreatePixInput{PlanoID: 2, MAC: "aa:bb:cc:dd:ee:05"})

	req := httptest.NewRequest("GET", "/admin/pagamentos/relatorio?formato=pdf", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "application/pdf") {
		t.Fatalf("content-type = %q", contentType)
	}
	if !strings.HasPrefix(string(body), "%PDF-1.4") {
		t.Fatalf("pdf invalido: %.20q", string(body))
	}
}

func TestPagamentosHandler_ValidaFiltros(t *testing.T) {
	app, _ := newPagamentosTestApp()

	cases := []string{
		"/admin/pagamentos?status=pago",
		"/admin/pagamentos?inicio=21-05-2026",
		"/admin/pagamentos?fim=amanha",
	}
	for _, target := range cases {
		t.Run(target, func(t *testing.T) {
			req := httptest.NewRequest("GET", target, nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 400 {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
			}
		})
	}
}

func newPagamentosTestApp() (*fiber.App, *memory.Store) {
	repo := memory.NewStore()
	app := fiber.New()
	deps := Dependencies{Config: config.Config{}, Store: repo}
	app.Get("/admin/pagamentos", pagamentosHandler(deps))
	app.Get("/admin/pagamentos/export.csv", exportPagamentosCSVHandler(deps))
	app.Get("/admin/pagamentos/relatorio", pagamentosRelatorioHandler(deps))
	return app, repo
}

func mustCreatePix(t *testing.T, repo *memory.Store, input store.CreatePixInput) store.PixTransaction {
	t.Helper()
	tx, err := repo.CreatePix(t.Context(), input)
	if err != nil {
		t.Fatal(err)
	}
	return tx
}
