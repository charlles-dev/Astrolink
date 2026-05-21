package admin

import (
	"github.com/astrolink/node/internal/config"
	"github.com/astrolink/node/internal/gateway"
	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

type Dependencies struct {
	Config  config.Config
	Store   store.Store
	Gateway gateway.Controller
}

func Register(app *fiber.App, deps Dependencies) {
	gatewayController := deps.Gateway
	if gatewayController == nil {
		gatewayController = gateway.NoopController{}
	}

	app.Post("/admin/auth/login", loginHandler(deps))
	app.Get("/admin/sistema/saude", healthHandler(deps))
	app.Get("/admin/planos", planosHandler(deps))
	app.Get("/admin/usuarios", usuariosHandler(deps))
	app.Get("/admin/vouchers", vouchersHandler(deps))
	app.Post("/admin/vouchers/gerar", gerarVouchersHandler(deps))
	app.Post("/admin/usuarios/:mac/desconectar", desconectarUsuarioHandler(gatewayController))
}

func adminError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{"erro": code, "mensagem": message})
}
