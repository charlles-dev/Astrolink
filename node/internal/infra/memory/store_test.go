package memory_test

import (
	"context"
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
