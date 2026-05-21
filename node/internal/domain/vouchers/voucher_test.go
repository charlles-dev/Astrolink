package vouchers_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/astrolink/node/internal/domain/vouchers"
)

func TestVoucher_Validar(t *testing.T) {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	max := 2
	expired := now.Add(-time.Hour)

	tests := []struct {
		name    string
		voucher vouchers.Voucher
		want    error
	}{
		{
			name:    "voucher_valido",
			voucher: vouchers.Voucher{Codigo: "ABCD-1234", Tipo: vouchers.TipoSingleUse, Ativo: true},
		},
		{
			name:    "voucher_ja_usado",
			voucher: vouchers.Voucher{Codigo: "ABCD-1234", Tipo: vouchers.TipoSingleUse, UsosAtuais: 1, Ativo: true},
			want:    vouchers.ErrJaUtilizado,
		},
		{
			name:    "voucher_universal_ainda_disponivel",
			voucher: vouchers.Voucher{Codigo: "UNIV-0000", Tipo: vouchers.TipoUniversal, UsosMaximos: &max, UsosAtuais: 1, Ativo: true},
		},
		{
			name:    "voucher_expirado",
			voucher: vouchers.Voucher{Codigo: "ABCD-1234", Tipo: vouchers.TipoSingleUse, ValidadeEm: &expired, Ativo: true},
			want:    vouchers.ErrExpirado,
		},
		{
			name:    "voucher_inativo",
			voucher: vouchers.Voucher{Codigo: "ABCD-1234", Tipo: vouchers.TipoSingleUse, Ativo: false},
			want:    vouchers.ErrInativo,
		},
		{
			name:    "codigo_vazio",
			voucher: vouchers.Voucher{Tipo: vouchers.TipoSingleUse, Ativo: true},
			want:    vouchers.ErrCodigoVazio,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.voucher.Validar(now)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Validar() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestGerarCodigo(t *testing.T) {
	codigo := vouchers.GerarCodigo("vip")
	if !strings.HasPrefix(codigo, "VIP") {
		t.Fatalf("codigo deve comecar com prefixo sanitizado VIP, got %q", codigo)
	}
	if !strings.Contains(codigo, "-") {
		t.Fatalf("codigo deve conter separador, got %q", codigo)
	}

	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		generated := vouchers.GerarCodigo("")
		if seen[generated] {
			t.Fatalf("codigo duplicado gerado: %q", generated)
		}
		seen[generated] = true
	}
}
