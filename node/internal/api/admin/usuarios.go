package admin

import (
	"errors"

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

func usuarioDetalheHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userStore, ok := deps.Store.(store.AdminUsuarioStore)
		if !ok {
			usuario, err := deps.Store.SessaoStatus(c.UserContext(), c.Params("mac"))
			if err != nil {
				return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar usuario")
			}
			if usuario.ID == 0 && usuario.Status == "walled_garden" {
				return adminError(c, fiber.StatusNotFound, "nao_encontrado", "usuario nao encontrado")
			}
			return c.JSON(fiber.Map{"usuario": usuario, "historico_sessoes": []store.UsuarioSessao{}, "total_sessoes": 0, "total_gasto": "0.00"})
		}
		detail, found, err := userStore.UsuarioByMAC(c.UserContext(), c.Params("mac"))
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar usuario")
		}
		if !found {
			return adminError(c, fiber.StatusNotFound, "nao_encontrado", "usuario nao encontrado")
		}
		return c.JSON(detail)
	}
}

func estenderUsuarioHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userStore, ok := deps.Store.(store.AdminUsuarioStore)
		if !ok {
			return adminError(c, fiber.StatusNotImplemented, "indisponivel", "operacao de usuario indisponivel")
		}
		var body struct {
			Minutos int `json:"minutos"`
		}
		if err := c.BodyParser(&body); err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "JSON invalido")
		}
		usuario, err := userStore.ExtendUsuario(c.UserContext(), store.ExtendUsuarioInput{
			MAC:     c.Params("mac"),
			Minutos: body.Minutos,
		})
		if errors.Is(err, store.ErrUsuarioNotFound) {
			return adminError(c, fiber.StatusNotFound, "nao_encontrado", "usuario nao encontrado")
		}
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", err.Error())
		}
		appendAdminLog(c.UserContext(), deps, store.AdminLogInput{
			Nivel:          "info",
			Tipo:           "usuarios",
			Mensagem:       "tempo de usuario estendido",
			MACRelacionado: usuario.MAC,
			Detalhes:       adminLogDetails(map[string]any{"mac": usuario.MAC, "minutos": body.Minutos}),
		})
		return c.JSON(fiber.Map{"usuario": usuario})
	}
}

func banirUsuarioHandler(deps Dependencies, gatewayController gateway.Controller) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userStore, ok := deps.Store.(store.AdminUsuarioStore)
		if !ok {
			return adminError(c, fiber.StatusNotImplemented, "indisponivel", "operacao de usuario indisponivel")
		}
		var body struct {
			Motivo string `json:"motivo"`
		}
		if err := c.BodyParser(&body); err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "JSON invalido")
		}
		usuario, err := userStore.BanUsuario(c.UserContext(), store.BanUsuarioInput{
			MAC:    c.Params("mac"),
			Motivo: body.Motivo,
		})
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", err.Error())
		}
		_ = gatewayController.Deauthorize(c.UserContext(), usuario.MAC)
		appendAdminLog(c.UserContext(), deps, store.AdminLogInput{
			Nivel:          "aviso",
			Tipo:           "usuarios",
			Mensagem:       "usuario banido",
			MACRelacionado: usuario.MAC,
			Detalhes:       adminLogDetails(map[string]any{"mac": usuario.MAC, "motivo": body.Motivo}),
		})
		return c.JSON(fiber.Map{"usuario": usuario})
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
