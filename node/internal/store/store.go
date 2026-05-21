package store

import (
	"context"
	"errors"
	"time"

	"github.com/astrolink/node/internal/domain/planos"
)

var (
	ErrPlanoNotFound   = errors.New("plano nao encontrado")
	ErrVoucherNotFound = errors.New("voucher nao encontrado")
	ErrInvalidQuantity = errors.New("quantidade invalida")
)

type Store interface {
	Settings(context.Context) (Settings, error)
	PortalPlanos(context.Context) ([]planos.Plano, error)
	AdminPlanos(context.Context) ([]planos.Plano, error)
	Usuarios(context.Context) ([]Usuario, error)
	SessaoStatus(context.Context, string) (Usuario, error)
	CreatePix(context.Context, CreatePixInput) (PixTransaction, error)
	PixStatus(context.Context, string) (PixTransaction, bool, error)
	RedeemVoucher(context.Context, RedeemVoucherInput) (RedeemVoucherResult, error)
	AdminVouchers(context.Context) ([]AdminVoucher, error)
	GenerateVouchers(context.Context, GenerateVouchersInput) (GenerateVouchersResult, error)
	Health(context.Context) Health
}

type AdminVoucherOperationalStore interface {
	AdminVouchersFiltered(context.Context, AdminVoucherFilter) ([]AdminVoucher, error)
	DeactivateVoucher(context.Context, int) (AdminVoucher, error)
}

type AdminAuthStore interface {
	CreateAdminSession(context.Context, CreateAdminSessionInput) error
	RotateAdminSession(context.Context, RotateAdminSessionInput) (AdminSession, bool, error)
	RevokeAdminSession(context.Context, string) error
}

type AdminPlanosStore interface {
	CreateAdminPlano(context.Context, AdminPlanoInput) (planos.Plano, error)
	UpdateAdminPlano(context.Context, int, AdminPlanoInput) (planos.Plano, error)
	SetAdminPlanoStatus(context.Context, int, bool) (planos.Plano, error)
}

type Settings struct {
	HotspotNome        string `json:"hotspot_nome"`
	HotspotLogoURL     string `json:"hotspot_logo_url"`
	CorPrimaria        string `json:"cor_primaria"`
	CorSecundaria      string `json:"cor_secundaria"`
	CorFundo           string `json:"cor_fundo"`
	MensagemBoasVindas string `json:"mensagem_boas_vindas"`
	URLPosConexao      string `json:"url_pos_conexao"`
	ColetaNome         bool   `json:"coleta_nome"`
	MostrarVelocidade  bool   `json:"mostrar_velocidade"`
}

func DefaultSettings() Settings {
	return Settings{
		HotspotNome:        "Astrolink Wi-Fi",
		HotspotLogoURL:     "",
		CorPrimaria:        "#06B6D4",
		CorSecundaria:      "#0E7490",
		CorFundo:           "#0F172A",
		MensagemBoasVindas: "Bem-vindo! Conecte-se e aproveite.",
		URLPosConexao:      "https://google.com",
		ColetaNome:         false,
		MostrarVelocidade:  true,
	}
}

type Usuario struct {
	ID                    int             `json:"id"`
	MAC                   string          `json:"mac"`
	IPAtual               string          `json:"ip_atual,omitempty"`
	Nome                  string          `json:"nome,omitempty"`
	Status                string          `json:"status"`
	Plano                 *PlanoResumo    `json:"plano,omitempty"`
	InicioAcesso          *time.Time      `json:"inicio_acesso,omitempty"`
	FimAcesso             *time.Time      `json:"fim_acesso,omitempty"`
	TempoRestanteSegundos int64           `json:"tempo_restante_segundos"`
	DadosConsumidosMB     int             `json:"dados_consumidos_mb"`
	Roteador              *RoteadorResumo `json:"roteador,omitempty"`
}

type PlanoResumo struct {
	ID   int    `json:"id"`
	Nome string `json:"nome"`
}

type RoteadorResumo struct {
	ID   int    `json:"id"`
	Nome string `json:"nome"`
}

type CreatePixInput struct {
	PlanoID int
	MAC     string
	IP      string
	Nome    string
}

type PixTransaction struct {
	TXID             string    `json:"txid"`
	Valor            string    `json:"valor"`
	Descricao        string    `json:"descricao"`
	PixCopiaCola     string    `json:"pix_copia_cola"`
	QRCodeBase64     string    `json:"qr_code_base64"`
	ExpiraEm         time.Time `json:"expira_em"`
	ExpiraEmSegundos int       `json:"expira_em_segundos"`
	Status           string    `json:"status,omitempty"`
	MAC              string    `json:"-"`
	PlanoID          int       `json:"-"`
}

type RedeemVoucherInput struct {
	Codigo string
	MAC    string
	IP     string
}

type RedeemVoucherResult struct {
	Usuario   Usuario
	Plano     planos.Plano
	HadAccess bool
}

type AdminVoucher struct {
	ID          int         `json:"id"`
	Codigo      string      `json:"codigo"`
	Plano       PlanoResumo `json:"plano"`
	Tipo        string      `json:"tipo"`
	UsosMaximos *int        `json:"usos_maximos"`
	UsosAtuais  int         `json:"usos_atuais"`
	ValidadeEm  *time.Time  `json:"validade_em"`
	Ativo       bool        `json:"ativo"`
	Prefixo     string      `json:"prefixo,omitempty"`
	LoteID      *int        `json:"lote_id,omitempty"`
	CreatedAt   time.Time   `json:"created_at,omitempty"`
}

type AdminVoucherFilter struct {
	Status  string
	PlanoID *int
	Codigo  string
	LoteID  *int
	Limit   int
}

type GenerateVouchersInput struct {
	PlanoID      int
	Quantidade   int
	Tipo         string
	UsosMaximos  *int
	ValidadeDias *int
	Prefixo      string
}

type GenerateVouchersResult struct {
	LoteID     int            `json:"lote_id"`
	Quantidade int            `json:"quantidade"`
	Vouchers   []AdminVoucher `json:"vouchers"`
}

type AdminPlanoInput struct {
	Nome           string
	Descricao      string
	Preco          float64
	DuracaoMinutos *int
	DadosMB        *int
	VelocidadeDown int
	VelocidadeUp   int
	Recomendado    bool
	Ativo          bool
	VisivelPortal  bool
	Ordem          int
}

type AdminSession struct {
	ID               string
	Usuario          string
	RefreshTokenHash string
	IP               string
	UserAgent        string
	ExpiresAt        time.Time
	Revoked          bool
	CreatedAt        time.Time
}

type CreateAdminSessionInput struct {
	Usuario          string
	RefreshTokenHash string
	IP               string
	UserAgent        string
	ExpiresAt        time.Time
}

type RotateAdminSessionInput struct {
	CurrentRefreshTokenHash string
	NextRefreshTokenHash    string
	IP                      string
	UserAgent               string
	ExpiresAt               time.Time
	Now                     time.Time
}

type Health struct {
	DatabaseStatus    string
	DatabaseLatencyMS int64
}
