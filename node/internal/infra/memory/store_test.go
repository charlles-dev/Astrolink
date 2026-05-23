package memory_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/astrolink/node/internal/infra/memory"
	"github.com/astrolink/node/internal/store"
)

func TestStore_UpdatePixStatus_AtualizaTransacaoExistente(t *testing.T) {
	repo := memory.NewStore()
	tx, err := repo.CreatePix(context.Background(), store.CreatePixInput{
		PlanoID: 2,
		MAC:     "AA:BB:CC:DD:EE:FF",
	})
	if err != nil {
		t.Fatalf("CreatePix() error = %v", err)
	}

	updated, ok, err := repo.UpdatePixStatus(context.Background(), store.UpdatePixStatusInput{
		TXID:   tx.TXID,
		Status: "aprovado",
	})
	if err != nil {
		t.Fatalf("UpdatePixStatus() error = %v", err)
	}
	if !ok {
		t.Fatal("UpdatePixStatus() ok = false, want true")
	}
	if updated.Status != "aprovado" {
		t.Fatalf("Status = %q, want aprovado", updated.Status)
	}
}

func TestStore_AppendAdminLog_AdminLogsRetornaEventoFiltravel(t *testing.T) {
	repo := memory.NewStore()
	ctx := context.Background()

	err := repo.AppendAdminLog(ctx, store.AdminLogInput{
		Nivel:    "info",
		Tipo:     "vouchers",
		Mensagem: "vouchers gerados",
		Detalhes: json.RawMessage(`{"lote_id":7,"quantidade":2}`),
	})
	if err != nil {
		t.Fatalf("AppendAdminLog() error = %v", err)
	}

	got, err := repo.AdminLogs(ctx, store.AdminLogFilter{
		Tipo:  "vouchers",
		Texto: "lote_id",
	})
	if err != nil {
		t.Fatalf("AdminLogs() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("AdminLogs() len = %d, want 1", len(got))
	}
	if got[0].Nivel != "info" || got[0].Tipo != "vouchers" || got[0].Mensagem != "vouchers gerados" {
		t.Fatalf("log inesperado: %+v", got[0])
	}
	if string(got[0].Detalhes) != `{"lote_id":7,"quantidade":2}` {
		t.Fatalf("Detalhes = %s", string(got[0].Detalhes))
	}
}

func TestStore_AdminLoginLockout_BloqueiaFalhasRecentesELiberaAposJanela(t *testing.T) {
	repo := memory.NewStore()
	ctx := context.Background()
	identity := store.AdminLoginIdentity{Usuario: "admin", IP: "192.0.2.10"}
	start := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)

	for i := 0; i < 4; i++ {
		status, err := repo.RecordAdminLoginFailure(ctx, store.AdminLoginFailureInput{
			Identity: identity,
			At:       start.Add(time.Duration(i) * time.Minute),
			Window:   15 * time.Minute,
			Limit:    5,
		})
		if err != nil {
			t.Fatalf("RecordAdminLoginFailure() error = %v", err)
		}
		if status.Locked {
			t.Fatalf("falha %d Locked = true, want false", i+1)
		}
	}

	status, err := repo.RecordAdminLoginFailure(ctx, store.AdminLoginFailureInput{
		Identity: identity,
		At:       start.Add(4 * time.Minute),
		Window:   15 * time.Minute,
		Limit:    5,
	})
	if err != nil {
		t.Fatalf("RecordAdminLoginFailure() error = %v", err)
	}
	if !status.Locked || status.Failures != 5 {
		t.Fatalf("status = %+v, want locked with 5 failures", status)
	}

	locked, err := repo.AdminLoginLocked(ctx, store.AdminLoginLockoutQuery{
		Identity: identity,
		Since:    start.Add(20 * time.Minute).Add(-15 * time.Minute),
		Limit:    5,
	})
	if err != nil {
		t.Fatalf("AdminLoginLocked() error = %v", err)
	}
	if locked {
		t.Fatal("AdminLoginLocked() apos janela = true, want false")
	}
}

func TestStore_AdminLoginLockout_LoginCorretoLimpaFalhas(t *testing.T) {
	repo := memory.NewStore()
	ctx := context.Background()
	identity := store.AdminLoginIdentity{Usuario: "admin", IP: "192.0.2.10"}
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)

	for i := 0; i < 4; i++ {
		if _, err := repo.RecordAdminLoginFailure(ctx, store.AdminLoginFailureInput{
			Identity: identity,
			At:       now.Add(time.Duration(i) * time.Minute),
			Window:   15 * time.Minute,
			Limit:    5,
		}); err != nil {
			t.Fatalf("RecordAdminLoginFailure() error = %v", err)
		}
	}

	if err := repo.ClearAdminLoginFailures(ctx, identity); err != nil {
		t.Fatalf("ClearAdminLoginFailures() error = %v", err)
	}
	locked, err := repo.AdminLoginLocked(ctx, store.AdminLoginLockoutQuery{
		Identity: identity,
		Since:    now.Add(-15 * time.Minute),
		Limit:    5,
	})
	if err != nil {
		t.Fatalf("AdminLoginLocked() error = %v", err)
	}
	if locked {
		t.Fatal("AdminLoginLocked() depois de limpar = true, want false")
	}
}
