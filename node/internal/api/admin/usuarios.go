package admin

import (
	"github.com/astrolink/node/internal/gateway"
	"github.com/astrolink/node/internal/store"
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

func desconectarUsuarioHandler(deps Dependencies, gatewayController gateway.Controller) fiber.Handler {
	return func(c *fiber.Ctx) error {
		mac := c.Params("mac")
		if err := gatewayController.Deauthorize(c.UserContext(), mac); err != nil {
			return adminError(c, fiber.StatusBadGateway, "roteador_indisponivel", "erro ao desconectar usuario no roteador")
		}
		appendAdminLog(c.UserContext(), deps, store.AdminLogInput{
			Nivel:    "info",
			Tipo:     "usuarios",
			Mensagem: "usuario desconectado",
			Detalhes: adminLogDetails(map[string]any{"mac": mac}),
		})
		return c.JSON(fiber.Map{"sucesso": true})
	}
}
