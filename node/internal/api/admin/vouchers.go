package admin

import (
	"errors"

	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

func vouchersHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		vouchers, err := deps.Store.AdminVouchers(c.UserContext())
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar vouchers")
		}
		return c.JSON(fiber.Map{
			"total":    len(vouchers),
			"vouchers": vouchers,
		})
	}
}

func gerarVouchersHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			PlanoID      int    `json:"plano_id"`
			Quantidade   int    `json:"quantidade"`
			Tipo         string `json:"tipo"`
			UsosMaximos  *int   `json:"usos_maximos"`
			ValidadeDias *int   `json:"validade_dias"`
			Prefixo      string `json:"prefixo"`
		}
		if err := c.BodyParser(&body); err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "JSON invalido")
		}
		result, err := deps.Store.GenerateVouchers(c.UserContext(), store.GenerateVouchersInput{
			PlanoID:      body.PlanoID,
			Quantidade:   body.Quantidade,
			Tipo:         body.Tipo,
			UsosMaximos:  body.UsosMaximos,
			ValidadeDias: body.ValidadeDias,
			Prefixo:      body.Prefixo,
		})
		if err != nil {
			return voucherAdminError(c, err)
		}
		return c.Status(fiber.StatusCreated).JSON(result)
	}
}

func voucherAdminError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, store.ErrPlanoNotFound):
		return adminError(c, fiber.StatusNotFound, "nao_encontrado", "plano nao encontrado")
	case errors.Is(err, store.ErrInvalidQuantity):
		return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "quantidade invalida")
	default:
		return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao gerar vouchers")
	}
}
