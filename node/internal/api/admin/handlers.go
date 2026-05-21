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
	app.Post("/admin/auth/refresh", refreshHandler(deps))

	protected := app.Group("/admin", authMiddleware(deps))
	protected.Post("/auth/logout", logoutHandler(deps))
	protected.Get("/auth/me", meHandler())
	protected.Get("/sistema/saude", healthHandler(deps))
	protected.Get("/planos", planosHandler(deps))
	protected.Get("/usuarios", usuariosHandler(deps))
	protected.Get("/vouchers", vouchersHandler(deps))
	protected.Post("/vouchers/gerar", gerarVouchersHandler(deps))
	protected.Post("/usuarios/:mac/desconectar", desconectarUsuarioHandler(gatewayController))
}

func adminError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{"erro": code, "mensagem": message})
}
