package admin

import (
	"context"
	"strconv"
	"time"

	"github.com/astrolink/node/internal/gateway"
	"github.com/gofiber/fiber/v2"
)

const (
	defaultRouterID   = 1
	defaultRouterName = "Roteador Local"
	routerPingTimeout = 300 * time.Millisecond
)

type localRouterStatus struct {
	status    string
	online    int
	offline   int
	latencyMS int64
}

func roteadoresHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		status := routerStatus(c.UserContext(), deps.Gateway)
		router := fiber.Map{
			"id":     defaultRouterID,
			"nome":   defaultRouterName,
			"status": status.status,
		}
		if status.latencyMS > 0 {
			router["latencia_ms"] = status.latencyMS
		}
		return c.JSON(fiber.Map{"roteadores": []fiber.Map{router}})
	}
}

func roteadorDiagnosticoHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil || id != defaultRouterID {
			return adminError(c, fiber.StatusNotFound, "roteador_nao_encontrado", "roteador nao encontrado")
		}

		controller, ok := deps.Gateway.(gateway.DiagnosticController)
		if !ok || isNoopGateway(deps.Gateway) {
			return c.JSON(fiber.Map{
				"status":      "dev/disabled",
				"roteador":    defaultRouterPayload("dev/disabled", 0),
				"diagnostico": gateway.RouterDiagnostic{},
			})
		}

		ctx, cancel := context.WithTimeout(c.UserContext(), routerPingTimeout)
		defer cancel()
		diagnostic, err := controller.Diagnostic(ctx)
		if err != nil {
			return c.JSON(fiber.Map{
				"status":      "offline",
				"roteador":    defaultRouterPayload("offline", 0),
				"erro":        "diagnostico_indisponivel",
				"diagnostico": gateway.RouterDiagnostic{},
			})
		}

		status := "offline"
		if diagnostic.Online {
			status = "online"
		}
		return c.JSON(fiber.Map{
			"status":      status,
			"roteador":    defaultRouterPayload(status, 0),
			"diagnostico": diagnostic,
		})
	}
}

func routerStatus(parent context.Context, controller gateway.Controller) localRouterStatus {
	if isNoopGateway(controller) {
		return localRouterStatus{status: "dev/disabled"}
	}

	ctx, cancel := context.WithTimeout(parent, routerPingTimeout)
	defer cancel()
	latency, err := controller.Ping(ctx)
	if err != nil {
		return localRouterStatus{status: "offline", offline: 1}
	}
	return localRouterStatus{
		status:    "online",
		online:    1,
		latencyMS: latency.Milliseconds(),
	}
}

func routerHealthPayload(status localRouterStatus) fiber.Map {
	payload := fiber.Map{
		"total":   1,
		"online":  status.online,
		"offline": status.offline,
		"status":  status.status,
	}
	if status.latencyMS > 0 {
		payload["latencia_ms"] = status.latencyMS
	}
	return payload
}

func defaultRouterPayload(status string, latencyMS int64) fiber.Map {
	payload := fiber.Map{
		"id":     defaultRouterID,
		"nome":   defaultRouterName,
		"status": status,
	}
	if latencyMS > 0 {
		payload["latencia_ms"] = latencyMS
	}
	return payload
}

func isNoopGateway(controller gateway.Controller) bool {
	if controller == nil {
		return true
	}
	switch controller.(type) {
	case gateway.NoopController, *gateway.NoopController:
		return true
	default:
		return false
	}
}
