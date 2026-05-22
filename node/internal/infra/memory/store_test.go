package memory_test

import (
	"context"
	"encoding/json"
	"testing"

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
