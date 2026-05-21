package admin

import (
	"encoding/base64"
	"time"

	"github.com/gofiber/fiber/v2"
)

func loginHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
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
	}
}
