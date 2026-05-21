package admin

import (
	"errors"
	"strconv"
	"strings"

	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

const maxPlanoOrdem = 9999

type adminPlanoBody struct {
	Nome           string  `json:"nome"`
	Descricao      string  `json:"descricao"`
	Preco          float64 `json:"preco"`
	DuracaoMinutos *int    `json:"duracao_minutos"`
	DadosMB        *int    `json:"dados_mb"`
	VelocidadeDown int     `json:"velocidade_down"`
	VelocidadeUp   int     `json:"velocidade_up"`
	Recomendado    bool    `json:"recomendado"`
	Ativo          *bool   `json:"ativo"`
	VisivelPortal  *bool   `json:"visivel_portal"`
	Ordem          int     `json:"ordem"`
}

func planosHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		planos, err := deps.Store.AdminPlanos(c.UserContext())
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar planos")
		}
		return c.JSON(fiber.Map{"planos": planos})
	}
}

func criarPlanoHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		input, err := parseAdminPlanoBody(c, true)
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", err.Error())
		}
		planosStore, ok := deps.Store.(store.AdminPlanosStore)
		if !ok {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "store de planos indisponivel")
		}
		plano, err := planosStore.CreateAdminPlano(c.UserContext(), input)
		if err != nil {
			return planoAdminError(c, err)
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"plano": plano})
	}
}

func atualizarPlanoHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := parsePlanoID(c)
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "id invalido")
		}
		input, err := parseAdminPlanoBody(c, false)
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", err.Error())
		}
		planosStore, ok := deps.Store.(store.AdminPlanosStore)
		if !ok {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "store de planos indisponivel")
		}
		plano, err := planosStore.UpdateAdminPlano(c.UserContext(), id, input)
		if err != nil {
			return planoAdminError(c, err)
		}
		return c.JSON(fiber.Map{"plano": plano})
	}
}

func alterarStatusPlanoHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := parsePlanoID(c)
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "id invalido")
		}
		var body struct {
			Ativo *bool `json:"ativo"`
		}
		if err := c.BodyParser(&body); err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "JSON invalido")
		}
		if body.Ativo == nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "ativo obrigatorio")
		}
		planosStore, ok := deps.Store.(store.AdminPlanosStore)
		if !ok {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "store de planos indisponivel")
		}
		plano, err := planosStore.SetAdminPlanoStatus(c.UserContext(), id, *body.Ativo)
		if err != nil {
			return planoAdminError(c, err)
		}
		return c.JSON(fiber.Map{"plano": plano})
	}
}

func parseAdminPlanoBody(c *fiber.Ctx, creating bool) (store.AdminPlanoInput, error) {
	var body adminPlanoBody
	if err := c.BodyParser(&body); err != nil {
		return store.AdminPlanoInput{}, errors.New("JSON invalido")
	}
	body.Nome = strings.TrimSpace(body.Nome)
	body.Descricao = strings.TrimSpace(body.Descricao)
	if body.Nome == "" {
		return store.AdminPlanoInput{}, errors.New("nome obrigatorio")
	}
	if body.Preco < 0 {
		return store.AdminPlanoInput{}, errors.New("preco deve ser maior ou igual a zero")
	}
	if body.DuracaoMinutos != nil && *body.DuracaoMinutos <= 0 {
		return store.AdminPlanoInput{}, errors.New("duracao_minutos deve ser positivo")
	}
	if body.DadosMB != nil && *body.DadosMB < 0 {
		return store.AdminPlanoInput{}, errors.New("dados_mb deve ser maior ou igual a zero")
	}
	if body.VelocidadeDown < 0 || body.VelocidadeUp < 0 {
		return store.AdminPlanoInput{}, errors.New("velocidades devem ser maiores ou iguais a zero")
	}
	if body.Ordem < 0 || body.Ordem > maxPlanoOrdem {
		return store.AdminPlanoInput{}, errors.New("ordem invalida")
	}
	ativo := false
	if body.Ativo != nil {
		ativo = *body.Ativo
	} else if creating {
		ativo = true
	}
	visivelPortal := false
	if body.VisivelPortal != nil {
		visivelPortal = *body.VisivelPortal
	} else if creating {
		visivelPortal = true
	}
	return store.AdminPlanoInput{
		Nome:           body.Nome,
		Descricao:      body.Descricao,
		Preco:          body.Preco,
		DuracaoMinutos: body.DuracaoMinutos,
		DadosMB:        body.DadosMB,
		VelocidadeDown: body.VelocidadeDown,
		VelocidadeUp:   body.VelocidadeUp,
		Recomendado:    body.Recomendado,
		Ativo:          ativo,
		VisivelPortal:  visivelPortal,
		Ordem:          body.Ordem,
	}, nil
}

func parsePlanoID(c *fiber.Ctx) (int, error) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return 0, errors.New("id invalido")
	}
	return id, nil
}

func planoAdminError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, store.ErrPlanoNotFound):
		return adminError(c, fiber.StatusNotFound, "nao_encontrado", "plano nao encontrado")
	default:
		return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao salvar plano")
	}
}
