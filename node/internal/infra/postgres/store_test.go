package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/astrolink/node/internal/infra/postgres"
	"github.com/astrolink/node/internal/store"
)

func TestStore_PortalPlanos_LerPlanosAtivosVisiveisOrdenados(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := postgres.NewStore(db, fixedClock)

	rows := sqlmock.NewRows([]string{
		"id", "nome", "descricao", "preco", "duracao_minutos", "dados_mb",
		"velocidade_down", "velocidade_up", "recomendado", "ativo", "visivel_portal", "ordem",
	}).AddRow(2, "Acesso 24 Horas", "Um dia completo", "15.00", 1440, nil, 10, 5, true, true, true, 1)

	mock.ExpectQuery("SELECT (.+) FROM planos").
		WillReturnRows(rows)

	got, err := repo.PortalPlanos(context.Background())
	if err != nil {
		t.Fatalf("PortalPlanos() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("PortalPlanos() len = %d, want 1", len(got))
	}
	if got[0].ID != 2 || got[0].Nome != "Acesso 24 Horas" || got[0].PrecoFormatado != "15.00" {
		t.Fatalf("plano mapeado incorretamente: %+v", got[0])
	}
	if got[0].DuracaoMinutos == nil || *got[0].DuracaoMinutos != 1440 || got[0].DuracaoFormatada != "24 horas" {
		t.Fatalf("duracao mapeada incorretamente: %+v", got[0])
	}
	assertExpectations(t, mock)
}

func TestStore_Settings_MesclaDefaultsComBanco(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := postgres.NewStore(db, fixedClock)

	rows := sqlmock.NewRows([]string{"chave", "valor"}).
		AddRow("hotspot_nome", "Wi-Fi Pousada").
		AddRow("cor_primaria", "#2ECC71").
		AddRow("coleta_nome", "true")
	mock.ExpectQuery("SELECT chave, valor FROM system_settings").
		WillReturnRows(rows)

	got, err := repo.Settings(context.Background())
	if err != nil {
		t.Fatalf("Settings() error = %v", err)
	}
	if got.HotspotNome != "Wi-Fi Pousada" {
		t.Fatalf("HotspotNome = %q, want Wi-Fi Pousada", got.HotspotNome)
	}
	if got.CorPrimaria != "#2ECC71" {
		t.Fatalf("CorPrimaria = %q, want #2ECC71", got.CorPrimaria)
	}
	if !got.ColetaNome {
		t.Fatal("ColetaNome = false, want true")
	}
	if got.URLPosConexao == "" {
		t.Fatal("URLPosConexao default deveria ser preservado")
	}
	assertExpectations(t, mock)
}

func TestStore_PixStatus_NaoEncontrado(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := postgres.NewStore(db, fixedClock)

	mock.ExpectQuery("SELECT (.+) FROM transacoes_pix").
		WithArgs("ast_missing").
		WillReturnError(sql.ErrNoRows)

	_, ok, err := repo.PixStatus(context.Background(), "ast_missing")
	if err != nil {
		t.Fatalf("PixStatus() error = %v", err)
	}
	if ok {
		t.Fatal("PixStatus() ok = true, want false")
	}
	assertExpectations(t, mock)
}

func TestStore_AdminPagamentos_FiltraEMapeiaPlano(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := postgres.NewStore(db, fixedClock)
	inicio := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	fim := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 5, 21, 3, 0, 0, 0, time.UTC)
	expiraEm := createdAt.Add(15 * time.Minute)

	rows := sqlmock.NewRows([]string{
		"txid", "status", "valor", "mac", "plano_id", "created_at", "expira_em", "id", "nome",
	}).AddRow("ast_123", "pendente", "15.00", "AA:BB:CC:DD:EE:FF", 2, createdAt, expiraEm, 2, "Acesso 24 Horas")

	mock.ExpectQuery("SELECT (.+) FROM transacoes_pix").
		WithArgs("pendente", inicio, fim).
		WillReturnRows(rows)

	got, err := repo.AdminPagamentos(context.Background(), store.AdminPagamentoFilter{
		Status:       "pendente",
		Inicio:       &inicio,
		Fim:          &fim,
		FimExclusive: true,
	})
	if err != nil {
		t.Fatalf("AdminPagamentos() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("AdminPagamentos() len = %d, want 1", len(got))
	}
	if got[0].TXID != "ast_123" || got[0].Plano.ID != 2 || got[0].Plano.Nome != "Acesso 24 Horas" {
		t.Fatalf("pagamento mapeado incorretamente: %+v", got[0])
	}
	if got[0].Descricao != "Astrolink Wi-Fi - Acesso 24 Horas" || !got[0].ExpiraEm.Equal(expiraEm) {
		t.Fatalf("descricao/datas incorretas: %+v", got[0])
	}
	assertExpectations(t, mock)
}

func TestStore_RedeemVoucher_CodigoInexistente(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := postgres.NewStore(db, fixedClock)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM vouchers").
		WithArgs("XXXX-9999").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err := repo.RedeemVoucher(context.Background(), store.RedeemVoucherInput{
		Codigo: "XXXX-9999",
		MAC:    "AA:BB:CC:DD:EE:FF",
		IP:     "192.168.1.50",
	})
	if !errors.Is(err, store.ErrVoucherNotFound) {
		t.Fatalf("RedeemVoucher() error = %v, want ErrVoucherNotFound", err)
	}
	assertExpectations(t, mock)
}

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	return db, mock
}

func fixedClock() time.Time {
	return time.Date(2026, 5, 21, 3, 0, 0, 0, time.UTC)
}

func assertExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
