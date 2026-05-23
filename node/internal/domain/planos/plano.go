package planos

import "fmt"

type Plano struct {
	ID               int     `json:"id"`
	Nome             string  `json:"nome"`
	Descricao        string  `json:"descricao,omitempty"`
	Preco            float64 `json:"-"`
	PrecoFormatado   string  `json:"preco"`
	DuracaoMinutos   *int    `json:"duracao_minutos"`
	DuracaoFormatada string  `json:"duracao_formatada"`
	DadosMB          *int    `json:"dados_mb"`
	VelocidadeDown   int     `json:"velocidade_down"`
	VelocidadeUp     int     `json:"velocidade_up"`
	Recomendado      bool    `json:"recomendado"`
	Ativo            bool    `json:"ativo"`
	VisivelPortal    bool    `json:"visivel_portal"`
	Ordem            int     `json:"ordem"`
}

type Config struct {
	ID             int
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

func New(id int, nome, descricao string, preco float64, duracaoMinutos *int, recomendado bool, ordem int) Plano {
	return FromConfig(Config{
		ID:             id,
		Nome:           nome,
		Descricao:      descricao,
		Preco:          preco,
		DuracaoMinutos: duracaoMinutos,
		VelocidadeDown: 10,
		VelocidadeUp:   5,
		Recomendado:    recomendado,
		Ativo:          true,
		VisivelPortal:  true,
		Ordem:          ordem,
	})
}

func FromConfig(config Config) Plano {
	return Plano{
		ID:               config.ID,
		Nome:             config.Nome,
		Descricao:        config.Descricao,
		Preco:            config.Preco,
		PrecoFormatado:   fmt.Sprintf("%.2f", config.Preco),
		DuracaoMinutos:   config.DuracaoMinutos,
		DuracaoFormatada: FormatDuration(config.DuracaoMinutos),
		DadosMB:          config.DadosMB,
		VelocidadeDown:   config.VelocidadeDown,
		VelocidadeUp:     config.VelocidadeUp,
		Recomendado:      config.Recomendado,
		Ativo:            config.Ativo,
		VisivelPortal:    config.VisivelPortal,
		Ordem:            config.Ordem,
	}
}

func FormatDuration(minutes *int) string {
	if minutes == nil {
		return "por dados"
	}
	if *minutes < 60 {
		if *minutes == 1 {
			return "1 minuto"
		}
		return fmt.Sprintf("%d minutos", *minutes)
	}
	if *minutes%1440 == 0 {
		days := *minutes / 1440
		if days == 1 {
			return "24 horas"
		}
		return fmt.Sprintf("%d dias", days)
	}
	if *minutes%60 == 0 {
		hours := *minutes / 60
		if hours == 1 {
			return "1 hora"
		}
		return fmt.Sprintf("%d horas", hours)
	}
	return fmt.Sprintf("%d minutos", *minutes)
}
