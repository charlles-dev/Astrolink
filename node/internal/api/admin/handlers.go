package admin

import (
	"encoding/base64"
	"errors"
	"time"

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

	app.Post("/admin/auth/login", func(c *fiber.Ctx) error {
		var body struct {
			Usuario string `json:"usuario"`
			Senha   string `json:"senha"`
		}
		if err := c.BodyParser(&body); err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "JSON invalido")
		}
		if body.Usuario != deps.Config.AdminUser || body.Senha != deps.Config.AdminPassword {
			return adminError(c, fiber.StatusUnauthorized, "nao_autenticado", "credenciais invalidas")
		}
		tokenPayload := body.Usuario + ":" + time.Now().UTC().Format(time.RFC3339)
		token := base64.RawURLEncoding.EncodeToString([]byte(tokenPayload))
		return c.JSON(fiber.Map{
			"access_token":  token,
			"refresh_token": token + ".refresh",
			"expires_in":    28800,
			"token_type":    "Bearer",
		})
	})

	app.Get("/admin/sistema/saude", func(c *fiber.Ctx) error {
		health := deps.Store.Health(c.UserContext())
		return c.JSON(fiber.Map{
			"status":          "healthy",
			"versao":          "0.1.0",
			"uptime_segundos": 0,
			"checks": fiber.Map{
				"banco_dados": fiber.Map{"status": health.DatabaseStatus, "latencia_ms": health.DatabaseLatencyMS},
				"redis":       fiber.Map{"status": "mock"},
				"rabbitmq":    fiber.Map{"status": "mock"},
				"mercadopago": fiber.Map{"status": "mock"},
				"roteadores":  fiber.Map{"total": 1, "online": 1, "offline": 0},
			},
		})
	})

	app.Get("/admin/planos", func(c *fiber.Ctx) error {
		planos, err := deps.Store.AdminPlanos(c.UserContext())
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar planos")
		}
		return c.JSON(fiber.Map{"planos": planos})
	})

	app.Get("/admin/usuarios", func(c *fiber.Ctx) error {
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
	})

	app.Get("/admin/vouchers", func(c *fiber.Ctx) error {
		vouchers, err := deps.Store.AdminVouchers(c.UserContext())
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar vouchers")
		}
		return c.JSON(fiber.Map{
			"total":    len(vouchers),
			"vouchers": vouchers,
		})
	})

	app.Post("/admin/vouchers/gerar", func(c *fiber.Ctx) error {
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
	})

	app.Post("/admin/usuarios/:mac/desconectar", func(c *fiber.Ctx) error {
		if err := gatewayController.Deauthorize(c.UserContext(), c.Params("mac")); err != nil {
			return adminError(c, fiber.StatusBadGateway, "roteador_indisponivel", "erro ao desconectar usuario no roteador")
		}
		return c.JSON(fiber.Map{"sucesso": true})
	})
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

func adminError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{"erro": code, "mensagem": message})
}
