package admin

import (
	"strings"
	"time"

	adminauth "github.com/astrolink/node/internal/auth"
	"github.com/gofiber/fiber/v2"
)

const adminUserLocalKey = "admin_usuario"

func authMiddleware(deps Dependencies) fiber.Handler {
	manager := adminauth.NewTokenManager(deps.Config.JWTSecret, time.Now)
	return func(c *fiber.Ctx) error {
		authorization := strings.TrimSpace(c.Get("Authorization"))
		if authorization == "" {
			return adminError(c, fiber.StatusUnauthorized, "nao_autenticado", "token ausente")
		}
		parts := strings.Fields(authorization)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return adminError(c, fiber.StatusUnauthorized, "nao_autenticado", "token invalido")
		}
		claims, err := manager.ValidateAccessToken(parts[1])
		if err != nil {
			return adminError(c, fiber.StatusUnauthorized, "nao_autenticado", "token invalido")
		}
		c.Locals(adminUserLocalKey, claims.Subject)
		return c.Next()
	}
}
