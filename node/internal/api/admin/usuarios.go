package admin

import (
	"github.com/astrolink/node/internal/gateway"
	"github.com/gofiber/fiber/v2"
)

func usuariosHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		usuarios, err := deps.Store.Usuarios(c.UserContext())
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar usuarios")
		}
		return c.JSON(fiber.Map{
			"total":    len(usuarios),
			"page":     1,
			"limit":    50,
			"usuarios": usuarios,
		})
	}
}

func desconectarUsuarioHandler(gatewayController gateway.Controller) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := gatewayController.Deauthorize(c.UserContext(), c.Params("mac")); err != nil {
			return adminError(c, fiber.StatusBadGateway, "roteador_indisponivel", "erro ao desconectar usuario no roteador")
		}
		return c.JSON(fiber.Map{"sucesso": true})
	}
}
