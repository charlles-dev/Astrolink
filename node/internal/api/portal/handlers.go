package portal

import (
	"bufio"
	"context"
	"errors"
	"time"

	"github.com/astrolink/node/internal/domain/vouchers"
	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App, appStore store.Store) {
	app.Get("/api/settings", func(c *fiber.Ctx) error {
		settings, err := appStore.Settings(requestContext(c))
		if err != nil {
			return apiError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar settings")
		}
		c.Set("Cache-Control", "public, max-age=300")
		return c.JSON(settings)
	})

	app.Get("/api/planos", func(c *fiber.Ctx) error {
		planos, err := appStore.PortalPlanos(requestContext(c))
		if err != nil {
			return apiError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar planos")
		}
		return c.JSON(fiber.Map{"planos": planos})
	})

	app.Get("/api/sessao/status", func(c *fiber.Ctx) error {
		usuario, err := appStore.SessaoStatus(requestContext(c), c.Query("mac"))
		if err != nil {
			return apiError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar sessao")
		}
		if usuario.Status != "ativo" {
			return c.JSON(fiber.Map{"ativa": false})
		}
		return c.JSON(fiber.Map{
			"ativa":                   true,
			"plano":                   usuario.Plano.Nome,
			"fim_acesso":              usuario.FimAcesso,
			"tempo_restante_segundos": usuario.TempoRestanteSegundos,
			"dados_consumidos_mb":     usuario.DadosConsumidosMB,
		})
	})

	app.Post("/api/pix/gerar", func(c *fiber.Ctx) error {
		var body struct {
			PlanoID int    `json:"plano_id"`
			MAC     string `json:"mac"`
			IP      string `json:"ip"`
			Nome    string `json:"nome"`
		}
		if err := c.BodyParser(&body); err != nil {
			return apiError(c, fiber.StatusBadRequest, "validacao_falhou", "JSON invalido")
		}
		tx, err := appStore.CreatePix(requestContext(c), store.CreatePixInput{
			PlanoID: body.PlanoID,
			MAC:     body.MAC,
			IP:      body.IP,
			Nome:    body.Nome,
		})
		if err != nil {
			return apiError(c, fiber.StatusBadRequest, "validacao_falhou", err.Error())
		}
		return c.Status(fiber.StatusCreated).JSON(tx)
	})

	app.Get("/api/pix/status/:txid", func(c *fiber.Ctx) error {
		tx, ok, err := appStore.PixStatus(requestContext(c), c.Params("txid"))
		if err != nil {
			return apiError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar PIX")
		}
		if !ok {
			return apiError(c, fiber.StatusNotFound, "nao_encontrado", "transacao PIX nao encontrada")
		}
		return c.JSON(fiber.Map{"txid": tx.TXID, "status": tx.Status, "expira_em": tx.ExpiraEm})
	})

	app.Get("/api/pix/aguardar/:txid", func(c *fiber.Ctx) error {
		txid := c.Params("txid")
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			for i := 0; i < 3; i++ {
				tx, ok, _ := appStore.PixStatus(context.Background(), txid)
				status := "nao_encontrado"
				if ok {
					status = tx.Status
				}
				_, _ = w.WriteString("event: status\n")
				_, _ = w.WriteString(`data: {"status":"` + status + `","txid":"` + txid + `"}` + "\n\n")
				_ = w.Flush()
				time.Sleep(time.Second)
			}
		})
		return nil
	})

	app.Post("/api/voucher/resgatar", func(c *fiber.Ctx) error {
		var body struct {
			Codigo string `json:"codigo"`
			MAC    string `json:"mac"`
			IP     string `json:"ip"`
		}
		if err := c.BodyParser(&body); err != nil {
			return apiError(c, fiber.StatusBadRequest, "validacao_falhou", "JSON invalido")
		}
		result, err := appStore.RedeemVoucher(requestContext(c), store.RedeemVoucherInput{
			Codigo: body.Codigo,
			MAC:    body.MAC,
			IP:     body.IP,
		})
		if err != nil {
			return voucherError(c, err)
		}
		minutes := 0
		if result.Plano.DuracaoMinutos != nil {
			minutes = *result.Plano.DuracaoMinutos
		}
		return c.JSON(fiber.Map{
			"sucesso":                  true,
			"plano":                    result.Plano.Nome,
			"tempo_adicionado_minutos": minutes,
			"fim_acesso":               result.Usuario.FimAcesso,
			"tempo_restante_segundos":  result.Usuario.TempoRestanteSegundos,
			"acesso_anterior":          result.HadAccess,
		})
	})
}

func voucherError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, store.ErrVoucherNotFound):
		return apiError(c, fiber.StatusNotFound, "nao_encontrado", "voucher nao encontrado")
	case errors.Is(err, vouchers.ErrJaUtilizado):
		return apiError(c, fiber.StatusGone, "recurso_esgotado", "voucher ja utilizado")
	case errors.Is(err, vouchers.ErrExpirado):
		return apiError(c, fiber.StatusUnprocessableEntity, "regra_negocio", "voucher expirado")
	case errors.Is(err, vouchers.ErrInativo):
		return apiError(c, fiber.StatusUnprocessableEntity, "regra_negocio", "voucher inativo")
	default:
		return apiError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao resgatar voucher")
	}
}

func apiError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{"erro": code, "mensagem": message})
}

func requestContext(c *fiber.Ctx) context.Context {
	if ctx := c.UserContext(); ctx != nil {
		return ctx
	}
	return context.Background()
}
