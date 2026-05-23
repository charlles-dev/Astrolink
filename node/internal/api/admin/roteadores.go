package admin

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/astrolink/node/internal/gateway"
	"github.com/astrolink/node/internal/store"
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
		if networkStore, ok := deps.Store.(store.AdminNetworkStore); ok {
			routers, err := networkStore.AdminRoteadores(c.UserContext())
			if err != nil {
				return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar roteadores")
			}
			if len(routers) > 0 {
				return c.JSON(fiber.Map{"roteadores": routers})
			}
		}
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

func criarRoteadorHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		networkStore, ok := deps.Store.(store.AdminNetworkStore)
		if !ok {
			return adminError(c, fiber.StatusNotImplemented, "indisponivel", "rede local indisponivel")
		}
		var body store.AdminRoteadorInput
		if err := c.BodyParser(&body); err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "JSON invalido")
		}
		router, err := networkStore.CreateRoteador(c.UserContext(), body)
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", err.Error())
		}
		appendAdminLog(c.UserContext(), deps, store.AdminLogInput{
			Nivel:    "info",
			Tipo:     "rede",
			Mensagem: "roteador criado",
			Detalhes: adminLogDetails(map[string]any{"id": router.ID, "nome": router.Nome, "ip": router.IP}),
		})
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"roteador": router})
	}
}

func atualizarRoteadorHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		networkStore, ok := deps.Store.(store.AdminNetworkStore)
		if !ok {
			return adminError(c, fiber.StatusNotImplemented, "indisponivel", "rede local indisponivel")
		}
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "id invalido")
		}
		var body store.AdminRoteadorInput
		if err := c.BodyParser(&body); err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "JSON invalido")
		}
		router, err := networkStore.UpdateRoteador(c.UserContext(), id, body)
		if errors.Is(err, store.ErrRouterNotFound) {
			return adminError(c, fiber.StatusNotFound, "nao_encontrado", "roteador nao encontrado")
		}
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", err.Error())
		}
		appendAdminLog(c.UserContext(), deps, store.AdminLogInput{
			Nivel:    "info",
			Tipo:     "rede",
			Mensagem: "roteador atualizado",
			Detalhes: adminLogDetails(map[string]any{"id": router.ID, "nome": router.Nome, "ip": router.IP}),
		})
		return c.JSON(fiber.Map{"roteador": router})
	}
}

func removerRoteadorHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		networkStore, ok := deps.Store.(store.AdminNetworkStore)
		if !ok {
			return adminError(c, fiber.StatusNotImplemented, "indisponivel", "rede local indisponivel")
		}
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "id invalido")
		}
		if err := networkStore.DeleteRoteador(c.UserContext(), id); errors.Is(err, store.ErrRouterNotFound) {
			return adminError(c, fiber.StatusNotFound, "nao_encontrado", "roteador nao encontrado")
		} else if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao remover roteador")
		}
		appendAdminLog(c.UserContext(), deps, store.AdminLogInput{
			Nivel:    "aviso",
			Tipo:     "rede",
			Mensagem: "roteador removido",
			Detalhes: adminLogDetails(map[string]any{"id": id}),
		})
		return c.SendStatus(fiber.StatusNoContent)
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

func roteadorSpeedtestHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil || id <= 0 {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "id invalido")
		}
		status := routerStatus(c.UserContext(), deps.Gateway)
		if status.status == "offline" {
			return adminError(c, fiber.StatusBadGateway, "roteador_indisponivel", "roteador offline")
		}
		return c.JSON(fiber.Map{
			"roteador_id":     id,
			"download_mbps":   0,
			"upload_mbps":     0,
			"status":          "indisponivel",
			"mensagem":        "speedtest real ainda depende do gateway OpenNDS ativo",
			"medido_em":       time.Now().UTC(),
			"roteador_status": status.status,
		})
	}
}

func blacklistHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		networkStore, ok := deps.Store.(store.AdminNetworkStore)
		if !ok {
			return adminError(c, fiber.StatusNotImplemented, "indisponivel", "rede local indisponivel")
		}
		items, err := networkStore.AdminBlacklist(c.UserContext())
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar blacklist")
		}
		return c.JSON(fiber.Map{"blacklist": items, "total": len(items)})
	}
}

func adicionarBlacklistHandler(deps Dependencies, gatewayController gateway.Controller) fiber.Handler {
	return func(c *fiber.Ctx) error {
		networkStore, ok := deps.Store.(store.AdminNetworkStore)
		if !ok {
			return adminError(c, fiber.StatusNotImplemented, "indisponivel", "rede local indisponivel")
		}
		var body store.AdminBlacklistInput
		if err := c.BodyParser(&body); err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "JSON invalido")
		}
		entry, err := networkStore.AddBlacklist(c.UserContext(), body)
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", err.Error())
		}
		_ = gatewayController.Deauthorize(c.UserContext(), entry.MAC)
		appendAdminLog(c.UserContext(), deps, store.AdminLogInput{
			Nivel:          "aviso",
			Tipo:           "rede",
			Mensagem:       "mac adicionado a blacklist",
			MACRelacionado: entry.MAC,
			Detalhes:       adminLogDetails(map[string]any{"mac": entry.MAC, "motivo": entry.Motivo}),
		})
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"entrada": entry})
	}
}

func removerBlacklistHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		networkStore, ok := deps.Store.(store.AdminNetworkStore)
		if !ok {
			return adminError(c, fiber.StatusNotImplemented, "indisponivel", "rede local indisponivel")
		}
		if err := networkStore.DeleteBlacklist(c.UserContext(), c.Params("mac")); err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao remover blacklist")
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

func walledGardenHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		networkStore, ok := deps.Store.(store.AdminNetworkStore)
		if !ok {
			return adminError(c, fiber.StatusNotImplemented, "indisponivel", "rede local indisponivel")
		}
		items, err := networkStore.AdminWalledGarden(c.UserContext())
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar walled garden")
		}
		return c.JSON(fiber.Map{"walled_garden": items, "total": len(items)})
	}
}

func adicionarWalledGardenHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		networkStore, ok := deps.Store.(store.AdminNetworkStore)
		if !ok {
			return adminError(c, fiber.StatusNotImplemented, "indisponivel", "rede local indisponivel")
		}
		var body store.AdminWalledGardenInput
		if err := c.BodyParser(&body); err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "JSON invalido")
		}
		entry, err := networkStore.AddWalledGarden(c.UserContext(), body)
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", err.Error())
		}
		appendAdminLog(c.UserContext(), deps, store.AdminLogInput{
			Nivel:    "info",
			Tipo:     "rede",
			Mensagem: "host adicionado ao walled garden",
			Detalhes: adminLogDetails(map[string]any{"host": entry.Host, "tipo": entry.Tipo}),
		})
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"entrada": entry})
	}
}

func removerWalledGardenHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		networkStore, ok := deps.Store.(store.AdminNetworkStore)
		if !ok {
			return adminError(c, fiber.StatusNotImplemented, "indisponivel", "rede local indisponivel")
		}
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "id invalido")
		}
		if err := networkStore.DeleteWalledGarden(c.UserContext(), id); err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", err.Error())
		}
		return c.SendStatus(fiber.StatusNoContent)
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
