package admin

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

type backupStore interface {
	CreateBackup(context.Context) (BackupResult, error)
}

type BackupResult struct {
	Arquivo   string `json:"arquivo,omitempty"`
	Tamanho   int64  `json:"tamanho_bytes,omitempty"`
	Mensagem  string `json:"mensagem,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

func backupHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		backupStore, ok := deps.Store.(backupStore)
		if !ok {
			return adminError(c, fiber.StatusNotImplemented, "backup_indisponivel", "backup manual e apenas para Postgres configurado")
		}
		result, err := backupStore.CreateBackup(c.UserContext())
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao criar backup")
		}
		return c.JSON(result)
	}
}
