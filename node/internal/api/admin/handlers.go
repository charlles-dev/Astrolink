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
	protected.Get("/setup/status", setupStatusHandler(deps))
	protected.Put("/setup/env", setupEnvHandler(deps))
	protected.Get("/planos", planosHandler(deps))
	protected.Post("/planos", criarPlanoHandler(deps))
	protected.Put("/planos/:id", atualizarPlanoHandler(deps))
	protected.Patch("/planos/:id/status", alterarStatusPlanoHandler(deps))
	protected.Get("/usuarios", usuariosHandler(deps))
	protected.Get("/usuarios/:mac", usuarioDetalheHandler(deps))
	protected.Post("/usuarios/:mac/estender", estenderUsuarioHandler(deps))
	protected.Post("/usuarios/:mac/banir", banirUsuarioHandler(deps, gatewayController))
	protected.Post("/usuarios/:mac/desconectar", desconectarUsuarioHandler(deps, gatewayController))
	protected.Get("/roteadores", roteadoresHandler(deps))
	protected.Post("/roteadores", criarRoteadorHandler(deps))
	protected.Put("/roteadores/:id", atualizarRoteadorHandler(deps))
	protected.Delete("/roteadores/:id", removerRoteadorHandler(deps))
	protected.Get("/roteadores/:id/diagnostico", roteadorDiagnosticoHandler(deps))
	protected.Post("/roteadores/:id/diagnostico", roteadorDiagnosticoHandler(deps))
	protected.Get("/rede/roteadores", roteadoresHandler(deps))
	protected.Post("/rede/roteadores", criarRoteadorHandler(deps))
	protected.Put("/rede/roteadores/:id", atualizarRoteadorHandler(deps))
	protected.Delete("/rede/roteadores/:id", removerRoteadorHandler(deps))
	protected.Post("/rede/roteadores/:id/diagnostico", roteadorDiagnosticoHandler(deps))
	protected.Post("/rede/roteadores/:id/speedtest", roteadorSpeedtestHandler(deps))
	protected.Get("/rede/blacklist", blacklistHandler(deps))
	protected.Post("/rede/blacklist", adicionarBlacklistHandler(deps, gatewayController))
	protected.Delete("/rede/blacklist/:mac", removerBlacklistHandler(deps))
	protected.Get("/rede/walled-garden", walledGardenHandler(deps))
	protected.Post("/rede/walled-garden", adicionarWalledGardenHandler(deps))
	protected.Delete("/rede/walled-garden/:id", removerWalledGardenHandler(deps))
	protected.Get("/pagamentos", pagamentosHandler(deps))
	protected.Get("/pagamentos/export.csv", exportPagamentosCSVHandler(deps))
	protected.Get("/pagamentos/relatorio", pagamentosRelatorioHandler(deps))
	protected.Get("/logs", logsHandler(deps))
	protected.Get("/logs/export.csv", exportLogsCSVHandler(deps))
	protected.Get("/sistema/logs", logsHandler(deps))
	protected.Get("/eventos", eventsHandler(deps))
	protected.Post("/backup", backupHandler(deps))
	protected.Post("/backup/restaurar", restoreBackupHandler(deps))
	protected.Get("/sistema/backup", backupHandler(deps))
	protected.Post("/sistema/restore", restoreBackupHandler(deps))
	protected.Get("/vouchers", vouchersHandler(deps))
	protected.Get("/vouchers/export.csv", exportVouchersCSVHandler(deps))
	protected.Post("/vouchers/gerar", gerarVouchersHandler(deps))
	protected.Patch("/vouchers/:id/desativar", desativarVoucherHandler(deps))
}

func adminError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{"erro": code, "mensagem": message})
}
