package payments

import (
	"context"
	"fmt"
	"time"
)

const (
	ProviderDemo = "demo"
)

type CreatePixInput struct {
	TXID      string
	Valor     string
	Descricao string
	ExpiresAt time.Time
}

type Pix struct {
	TXID         string
	PixCopiaCola string
	QRCodeBase64 string
	ExpiresAt    time.Time
}

type Provider interface {
	CreatePix(context.Context, CreatePixInput) (Pix, error)
}

func NewProvider(name string) Provider {
	switch name {
	case "", ProviderDemo:
		return DemoProvider{}
	default:
		return DemoProvider{}
	}
}

type DemoProvider struct{}

func (DemoProvider) CreatePix(_ context.Context, input CreatePixInput) (Pix, error) {
	if input.TXID == "" {
		return Pix{}, fmt.Errorf("txid obrigatorio")
	}
	return Pix{
		TXID:         input.TXID,
		PixCopiaCola: "00020126580014br.gov.bcb.pix0136astrolink-demo-" + input.TXID,
		QRCodeBase64: "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNTYiIGhlaWdodD0iMjU2Ij48cmVjdCB3aWR0aD0iMjU2IiBoZWlnaHQ9IjI1NiIgZmlsbD0id2hpdGUiLz48dGV4dCB4PSIxMjgiIHk9IjEyOCIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZmlsbD0iYmxhY2siPkFzdHJvbGluayBQSVg8L3RleHQ+PC9zdmc+",
		ExpiresAt:    input.ExpiresAt,
	}, nil
}
