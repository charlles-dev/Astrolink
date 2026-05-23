package admin

import (
	"strings"

	"github.com/astrolink/node/internal/config"
	"github.com/gofiber/fiber/v2"
)

type setupEnvRequest struct {
	Values map[string]string `json:"values"`
}

func setupStatusHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		status, err := loadSetupStatus(deps.Config)
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao ler configuracao local")
		}
		return c.JSON(status)
	}
}

func setupEnvHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !deps.Config.AstrolinkAllowEnvWrite {
			return adminError(c, fiber.StatusForbidden, "setup_escrita_desabilitada", "escrita do .env pelo painel desabilitada")
		}

		var body setupEnvRequest
		if err := c.BodyParser(&body); err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "JSON invalido")
		}
		if body.Values == nil {
			body.Values = map[string]string{}
		}

		envPath := setupEnvPath(deps.Config)
		file, err := config.LoadEnvFile(envPath)
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao ler arquivo .env")
		}
		if err := config.ApplySetupPatch(file, body.Values); err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "variavel de configuracao nao permitida")
		}
		if err := config.SaveEnvFileAtomic(envPath, file); err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao gravar arquivo .env")
		}

		status := config.BuildSetupStatus(file)
		status.Writable = true
		status.RequiresRestart = len(body.Values) > 0
		return c.JSON(status)
	}
}

func loadSetupStatus(cfg config.Config) (config.SetupStatus, error) {
	file, err := config.LoadEnvFile(setupEnvPath(cfg))
	if err != nil {
		return config.SetupStatus{}, err
	}
	status := config.BuildSetupStatus(file)
	status.Writable = cfg.AstrolinkAllowEnvWrite
	return status, nil
}

func setupEnvPath(cfg config.Config) string {
	path := strings.TrimSpace(cfg.AstrolinkEnvFile)
	if path == "" {
		return ".env"
	}
	return path
}
