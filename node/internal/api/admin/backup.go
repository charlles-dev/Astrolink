package admin

import (
	"context"
	"strings"

	"github.com/astrolink/node/internal/store"
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

type restoreBackupRequest struct {
	Arquivo     string `json:"arquivo"`
	Confirmacao string `json:"confirmacao"`
}

func backupHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		appendAdminLog(c.UserContext(), deps, store.AdminLogInput{
			Nivel:    "info",
			Tipo:     "backup",
			Mensagem: "backup solicitado",
		})
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

func restoreBackupHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body restoreBackupRequest
		if err := c.BodyParser(&body); err != nil {
			appendAdminLog(c.UserContext(), deps, store.AdminLogInput{
				Nivel:    "aviso",
				Tipo:     "restore",
				Mensagem: "restore bloqueado",
				Detalhes: adminLogDetails(map[string]any{
					"motivo": "json_invalido",
				}),
			})
			return adminError(c, fiber.StatusBadRequest, "requisicao_invalida", "informe arquivo e confirmacao RESTAURAR")
		}
		if strings.TrimSpace(body.Arquivo) == "" || body.Confirmacao != "RESTAURAR" {
			appendAdminLog(c.UserContext(), deps, store.AdminLogInput{
				Nivel:    "aviso",
				Tipo:     "restore",
				Mensagem: "restore bloqueado",
				Detalhes: adminLogDetails(map[string]any{
					"arquivo": strings.TrimSpace(body.Arquivo),
					"motivo":  "confirmacao_invalida",
				}),
			})
			return adminError(c, fiber.StatusBadRequest, "confirmacao_invalida", "para validar restore, informe arquivo e confirmacao RESTAURAR")
		}
		appendAdminLog(c.UserContext(), deps, store.AdminLogInput{
			Nivel:    "aviso",
			Tipo:     "restore",
			Mensagem: "restore validado e bloqueado",
			Detalhes: adminLogDetails(map[string]any{
				"arquivo": strings.TrimSpace(body.Arquivo),
				"motivo":  "procedimento_manual",
			}),
		})
		return adminError(c, fiber.StatusNotImplemented, "restore_indisponivel", "restore real exige procedimento manual com Postgres; nenhuma restauracao foi executada")
	}
}
