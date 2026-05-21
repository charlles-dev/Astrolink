package admin

import (
	"time"

	adminauth "github.com/astrolink/node/internal/auth"
	"github.com/astrolink/node/internal/store"
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
		authStore, ok := deps.Store.(store.AdminAuthStore)
		if !ok {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "store de auth indisponivel")
		}

		manager := adminauth.NewTokenManager(deps.Config.JWTSecret, time.Now)
		accessToken, _, err := manager.GenerateAccessToken(body.Usuario)
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao gerar token")
		}
		refreshToken, err := adminauth.NewRefreshToken()
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao gerar refresh token")
		}
		now := time.Now().UTC()
		err = authStore.CreateAdminSession(c.UserContext(), store.CreateAdminSessionInput{
			Usuario:          body.Usuario,
			RefreshTokenHash: adminauth.HashRefreshToken(refreshToken),
			IP:               c.IP(),
			UserAgent:        c.Get("User-Agent"),
			ExpiresAt:        now.Add(adminauth.RefreshTokenTTL),
		})
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao criar sessao")
		}
		return authResponse(c, accessToken, refreshToken)
	}
}

func refreshHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.BodyParser(&body); err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "JSON invalido")
		}
		if body.RefreshToken == "" {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "refresh_token obrigatorio")
		}
		authStore, ok := deps.Store.(store.AdminAuthStore)
		if !ok {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "store de auth indisponivel")
		}

		nextRefreshToken, err := adminauth.NewRefreshToken()
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao gerar refresh token")
		}
		now := time.Now().UTC()
		session, ok, err := authStore.RotateAdminSession(c.UserContext(), store.RotateAdminSessionInput{
			CurrentRefreshTokenHash: adminauth.HashRefreshToken(body.RefreshToken),
			NextRefreshTokenHash:    adminauth.HashRefreshToken(nextRefreshToken),
			IP:                      c.IP(),
			UserAgent:               c.Get("User-Agent"),
			ExpiresAt:               now.Add(adminauth.RefreshTokenTTL),
			Now:                     now,
		})
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao renovar sessao")
		}
		if !ok {
			return adminError(c, fiber.StatusUnauthorized, "nao_autenticado", "refresh token invalido")
		}

		manager := adminauth.NewTokenManager(deps.Config.JWTSecret, time.Now)
		accessToken, _, err := manager.GenerateAccessToken(session.Usuario)
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao gerar token")
		}
		return authResponse(c, accessToken, nextRefreshToken)
	}
}

func logoutHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.BodyParser(&body); err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "JSON invalido")
		}
		if body.RefreshToken == "" {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "refresh_token obrigatorio")
		}
		authStore, ok := deps.Store.(store.AdminAuthStore)
		if !ok {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "store de auth indisponivel")
		}
		if err := authStore.RevokeAdminSession(c.UserContext(), adminauth.HashRefreshToken(body.RefreshToken)); err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao encerrar sessao")
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

func meHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		usuario, _ := c.Locals(adminUserLocalKey).(string)
		return c.JSON(fiber.Map{"usuario": usuario})
	}
}

func authResponse(c *fiber.Ctx, accessToken, refreshToken string) error {
	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    int(adminauth.AccessTokenTTL.Seconds()),
		"token_type":    "Bearer",
	})
}
