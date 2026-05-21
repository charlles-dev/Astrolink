package vouchers

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"
)

var (
	ErrInativo     = errors.New("voucher inativo")
	ErrJaUtilizado = errors.New("voucher ja utilizado")
	ErrExpirado    = errors.New("voucher expirado")
	ErrCodigoVazio = errors.New("codigo vazio")
)

type Tipo string

const (
	TipoSingleUse Tipo = "single_use"
	TipoUniversal Tipo = "universal"
)

type Voucher struct {
	ID          int
	Codigo      string
	PlanoID     int
	Tipo        Tipo
	UsosMaximos *int
	UsosAtuais  int
	ValidadeEm  *time.Time
	Ativo       bool
	Prefixo     string
}

func (v Voucher) Validar(now time.Time) error {
	if strings.TrimSpace(v.Codigo) == "" {
		return ErrCodigoVazio
	}
	if !v.Ativo {
		return ErrInativo
	}
	if v.ValidadeEm != nil && now.After(*v.ValidadeEm) {
		return ErrExpirado
	}
	limit := 1
	if v.Tipo == TipoUniversal && v.UsosMaximos != nil {
		limit = *v.UsosMaximos
	}
	if v.UsosAtuais >= limit {
		return ErrJaUtilizado
	}
	return nil
}

func GerarCodigo(prefixo string) string {
	cleanPrefix := strings.ToUpper(strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, strings.ToUpper(prefixo)))
	if len(cleanPrefix) > 6 {
		cleanPrefix = cleanPrefix[:6]
	}
	return cleanPrefix + randomChunk(4) + "-" + randomChunk(4)
}

func randomChunk(size int) string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var builder strings.Builder
	builder.Grow(size)
	max := big.NewInt(int64(len(alphabet)))
	for builder.Len() < size {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			builder.WriteByte(alphabet[int(time.Now().UnixNano()%int64(len(alphabet)))])
			continue
		}
		builder.WriteByte(alphabet[int(n.Int64())])
	}
	return builder.String()
}
