package admin

import "github.com/gofiber/fiber/v2"

func healthHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		health := deps.Store.Health(c.UserContext())
		router := routerStatus(c.UserContext(), deps.Gateway)
		return c.JSON(fiber.Map{
			"status":          "healthy",
			"versao":          "0.1.0",
			"uptime_segundos": 0,
			"checks": fiber.Map{
				"banco_dados": fiber.Map{"status": health.DatabaseStatus, "latencia_ms": health.DatabaseLatencyMS},
				"redis":       fiber.Map{"status": "mock"},
				"rabbitmq":    fiber.Map{"status": "mock"},
				"mercadopago": fiber.Map{"status": "mock"},
				"roteadores":  routerHealthPayload(router),
			},
		})
	}
}
