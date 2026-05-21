package admin

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

func vouchersHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filter, err := adminVoucherFilterFromQuery(c, 200)
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", err.Error())
		}
		vouchers, err := adminVouchers(c, deps, filter)
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar vouchers")
		}
		return c.JSON(fiber.Map{
			"total":    len(vouchers),
			"vouchers": vouchers,
		})
	}
}

func desativarVoucherHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil || id <= 0 {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "id do voucher invalido")
		}
		voucherStore, ok := deps.Store.(store.AdminVoucherOperationalStore)
		if !ok {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "store de vouchers indisponivel")
		}
		voucher, err := voucherStore.DeactivateVoucher(c.UserContext(), id)
		if err != nil {
			return deactivateVoucherAdminError(c, err)
		}
		return c.JSON(fiber.Map{"voucher": voucher})
	}
}

func exportVouchersCSVHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filter, err := adminVoucherFilterFromQuery(c, 0)
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", err.Error())
		}
		vouchers, err := adminVouchers(c, deps, filter)
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao exportar vouchers")
		}
		var buffer bytes.Buffer
		writer := csv.NewWriter(&buffer)
		if err := writer.Write([]string{"codigo", "plano", "tipo", "usos_atuais", "usos_maximos", "ativo", "validade_em", "prefixo", "lote_id", "created_at"}); err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao exportar vouchers")
		}
		for _, voucher := range vouchers {
			if err := writer.Write(adminVoucherCSVRecord(voucher)); err != nil {
				return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao exportar vouchers")
			}
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao exportar vouchers")
		}
		c.Type("csv")
		c.Set(fiber.HeaderContentDisposition, `attachment; filename="vouchers.csv"`)
		return c.Send(buffer.Bytes())
	}
}

func gerarVouchersHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
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
	}
}

func adminVouchers(c *fiber.Ctx, deps Dependencies, filter store.AdminVoucherFilter) ([]store.AdminVoucher, error) {
	voucherStore, ok := deps.Store.(store.AdminVoucherOperationalStore)
	if ok {
		return voucherStore.AdminVouchersFiltered(c.UserContext(), filter)
	}
	return deps.Store.AdminVouchers(c.UserContext())
}

func adminVoucherFilterFromQuery(c *fiber.Ctx, limit int) (store.AdminVoucherFilter, error) {
	status := strings.ToLower(strings.TrimSpace(c.Query("status")))
	switch status {
	case "", "ativo", "inativo", "todos":
	default:
		return store.AdminVoucherFilter{}, errors.New("status invalido")
	}
	filter := store.AdminVoucherFilter{
		Status: status,
		Codigo: strings.TrimSpace(c.Query("codigo")),
		Limit:  limit,
	}
	if planoID, err := optionalPositiveIntQuery(c, "plano_id"); err != nil {
		return store.AdminVoucherFilter{}, err
	} else {
		filter.PlanoID = planoID
	}
	if loteID, err := optionalPositiveIntQuery(c, "lote_id"); err != nil {
		return store.AdminVoucherFilter{}, err
	} else {
		filter.LoteID = loteID
	}
	return filter, nil
}

func optionalPositiveIntQuery(c *fiber.Ctx, name string) (*int, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return nil, errors.New(name + " invalido")
	}
	return &value, nil
}

func adminVoucherCSVRecord(voucher store.AdminVoucher) []string {
	usosMaximos := ""
	if voucher.UsosMaximos != nil {
		usosMaximos = strconv.Itoa(*voucher.UsosMaximos)
	}
	validadeEm := formatCSVTime(voucher.ValidadeEm)
	loteID := ""
	if voucher.LoteID != nil {
		loteID = strconv.Itoa(*voucher.LoteID)
	}
	return []string{
		voucher.Codigo,
		voucher.Plano.Nome,
		voucher.Tipo,
		strconv.Itoa(voucher.UsosAtuais),
		usosMaximos,
		strconv.FormatBool(voucher.Ativo),
		validadeEm,
		voucher.Prefixo,
		loteID,
		formatCSVTime(&voucher.CreatedAt),
	}
}

func formatCSVTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
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

func deactivateVoucherAdminError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, store.ErrVoucherNotFound):
		return adminError(c, fiber.StatusNotFound, "nao_encontrado", "voucher nao encontrado")
	default:
		return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao desativar voucher")
	}
}
