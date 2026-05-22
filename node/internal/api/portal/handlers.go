package portal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/astrolink/node/internal/domain/planos"
	"github.com/astrolink/node/internal/domain/vouchers"
	"github.com/astrolink/node/internal/gateway"
	"github.com/astrolink/node/internal/payments"
	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

type Dependencies struct {
	Store                    store.Store
	Gateway                  gateway.Controller
	Logger                   *slog.Logger
	Env                      string
	MercadoPagoWebhookSecret string
	PaymentProvider          payments.Provider
}

func Register(app *fiber.App, deps Dependencies) {
	appStore := deps.Store
	gatewayController := deps.Gateway
	if gatewayController == nil {
		gatewayController = gateway.NoopController{}
	}
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	env := strings.ToLower(strings.TrimSpace(deps.Env))
	if env == "" {
		env = "development"
	}
	paymentProvider := deps.PaymentProvider
	if paymentProvider == nil {
		paymentProvider = payments.NewProvider(payments.ProviderConfig{Name: payments.ProviderDemo})
	}

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
		ctx := requestContext(c)
		plano, ok, err := findPortalPlano(ctx, appStore, body.PlanoID)
		if err != nil {
			return apiError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar plano")
		}
		if !ok {
			return apiError(c, fiber.StatusBadRequest, "validacao_falhou", store.ErrPlanoNotFound.Error())
		}
		now := time.Now().UTC()
		expiresAt := now.Add(15 * time.Minute)
		txid := newPortalPixTXID()
		description := "Astrolink Wi-Fi - " + plano.Nome
		pix, err := paymentProvider.CreatePix(ctx, payments.CreatePixInput{
			TXID:      txid,
			Valor:     plano.PrecoFormatado,
			Descricao: description,
			ExpiresAt: expiresAt,
		})
		if err != nil {
			return apiError(c, fiber.StatusBadGateway, "provider_indisponivel", "erro ao gerar PIX")
		}
		if strings.TrimSpace(pix.TXID) != "" {
			txid = strings.TrimSpace(pix.TXID)
		}
		if !pix.ExpiresAt.IsZero() {
			expiresAt = pix.ExpiresAt
		}
		tx, err := appStore.CreatePix(ctx, store.CreatePixInput{
			PlanoID:      body.PlanoID,
			MAC:          body.MAC,
			IP:           body.IP,
			Nome:         body.Nome,
			TXID:         txid,
			PixCopiaCola: pix.PixCopiaCola,
			QRCodeBase64: pix.QRCodeBase64,
			ExpiraEm:     expiresAt,
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

	app.Post("/api/pix/dev/aprovar/:txid", func(c *fiber.Ctx) error {
		if env != "development" {
			return apiError(c, fiber.StatusNotFound, "nao_encontrado", "rota nao encontrada")
		}
		tx, ok, err := appStore.UpdatePixStatus(requestContext(c), store.UpdatePixStatusInput{
			TXID:   c.Params("txid"),
			Status: string(payments.PixStatusApproved),
		})
		if err != nil {
			return apiError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao aprovar PIX")
		}
		if !ok {
			return apiError(c, fiber.StatusNotFound, "nao_encontrado", "transacao PIX nao encontrada")
		}
		return c.JSON(fiber.Map{"txid": tx.TXID, "status": tx.Status})
	})

	app.Post("/api/webhooks/mercadopago", func(c *fiber.Ctx) error {
		event := parseMercadoPagoWebhook(c)
		err := (payments.MercadoPagoWebhookVerifier{Secret: deps.MercadoPagoWebhookSecret}).Verify(payments.MercadoPagoWebhookVerification{
			DataID:    event.PaymentID,
			RequestID: c.Get("x-request-id"),
			Signature: c.Get("x-signature"),
		})
		if err != nil {
			if errors.Is(err, payments.ErrWebhookSecretMissing) && env != "production" {
				logger.Info("webhook Mercado Pago ignorado sem segredo configurado")
				return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "ignored", "reason": "webhook_secret_missing"})
			}
			return apiError(c, fiber.StatusUnauthorized, "assinatura_invalida", "webhook Mercado Pago nao autenticado")
		}

		result, err := paymentProvider.PixStatus(requestContext(c), payments.StatusQuery{
			PaymentID: event.PaymentID,
			TXID:      event.TXID,
		})
		if err != nil {
			return apiError(c, fiber.StatusBadGateway, "provider_indisponivel", "erro ao consultar Mercado Pago")
		}
		if result.Status != payments.PixStatusApproved {
			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "ignored"})
		}
		txid := result.TXID
		if txid == "" {
			txid = event.TXID
		}
		if txid == "" {
			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "ignored", "reason": "txid_missing"})
		}
		tx, ok, err := appStore.UpdatePixStatus(requestContext(c), store.UpdatePixStatusInput{
			TXID:   txid,
			Status: string(payments.PixStatusApproved),
		})
		if err != nil {
			return apiError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao atualizar PIX")
		}
		if !ok {
			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "ignored", "reason": "txid_not_found"})
		}
		return c.JSON(fiber.Map{"status": "ok", "txid": tx.TXID})
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
		minutes := 60
		if result.Plano.DuracaoMinutos != nil {
			minutes = *result.Plano.DuracaoMinutos
		}
		mac := result.Usuario.MAC
		if mac == "" {
			mac = body.MAC
		}
		routerAuthorized := true
		if err := gatewayController.Authorize(requestContext(c), gateway.Authorization{
			MAC:      mac,
			Duration: time.Duration(minutes) * time.Minute,
		}); err != nil {
			routerAuthorized = false
			logger.Warn("falha ao autorizar cliente no OpenNDS", "mac", mac, "error", err)
		}
		return c.JSON(fiber.Map{
			"sucesso":                  true,
			"plano":                    result.Plano.Nome,
			"tempo_adicionado_minutos": minutes,
			"fim_acesso":               result.Usuario.FimAcesso,
			"tempo_restante_segundos":  result.Usuario.TempoRestanteSegundos,
			"acesso_anterior":          result.HadAccess,
			"roteador_autorizado":      routerAuthorized,
		})
	})
}

type mercadoPagoWebhookEvent struct {
	PaymentID string
	TXID      string
}

func parseMercadoPagoWebhook(c *fiber.Ctx) mercadoPagoWebhookEvent {
	event := mercadoPagoWebhookEvent{
		PaymentID: strings.TrimSpace(firstNonEmpty(c.Query("data.id"), c.Query("id"))),
		TXID:      strings.TrimSpace(c.Query("txid")),
	}
	var body struct {
		ID   any    `json:"id"`
		TXID string `json:"txid"`
		Data struct {
			ID any `json:"id"`
		} `json:"data"`
		ExternalReference string `json:"external_reference"`
	}
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.UseNumber()
	if err := decoder.Decode(&body); err == nil {
		if event.PaymentID == "" {
			event.PaymentID = stringifyWebhookID(body.Data.ID)
		}
		if event.PaymentID == "" {
			event.PaymentID = stringifyWebhookID(body.ID)
		}
		if event.TXID == "" {
			event.TXID = strings.TrimSpace(firstNonEmpty(body.TXID, body.ExternalReference))
		}
	}
	return event
}

func stringifyWebhookID(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	case json.Number:
		return strings.TrimSpace(typed.String())
	default:
		return ""
	}
}

func findPortalPlano(ctx context.Context, appStore store.Store, id int) (planos.Plano, bool, error) {
	items, err := appStore.PortalPlanos(ctx)
	if err != nil {
		return planos.Plano{}, false, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, true, nil
		}
	}
	return planos.Plano{}, false, nil
}

func newPortalPixTXID() string {
	return "ast_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
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
