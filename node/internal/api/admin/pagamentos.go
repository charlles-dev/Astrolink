package admin

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

func pagamentosHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filter, err := adminPagamentoFilterFromQuery(c)
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", err.Error())
		}
		pagamentos, err := adminPagamentos(c, deps, filter)
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar pagamentos")
		}
		return c.JSON(fiber.Map{
			"total":      len(pagamentos),
			"totais":     adminPagamentoTotals(pagamentos),
			"pagamentos": pagamentos,
		})
	}
}

func exportPagamentosCSVHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filter, err := adminPagamentoFilterFromQuery(c)
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", err.Error())
		}
		pagamentos, err := adminPagamentos(c, deps, filter)
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao exportar pagamentos")
		}
		var buffer bytes.Buffer
		writer := csv.NewWriter(&buffer)
		if err := writer.Write([]string{"txid", "status", "valor", "descricao", "mac", "plano_id", "plano", "created_at", "expira_em"}); err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao exportar pagamentos")
		}
		for _, pagamento := range pagamentos {
			if err := writer.Write(adminPagamentoCSVRecord(pagamento)); err != nil {
				return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao exportar pagamentos")
			}
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao exportar pagamentos")
		}
		c.Type("csv")
		c.Set(fiber.HeaderContentDisposition, `attachment; filename="pagamentos.csv"`)
		return c.Send(buffer.Bytes())
	}
}

func adminPagamentos(c *fiber.Ctx, deps Dependencies, filter store.AdminPagamentoFilter) ([]store.AdminPagamento, error) {
	pagamentoStore, ok := deps.Store.(store.AdminPagamentoStore)
	if !ok {
		return nil, errors.New("store de pagamentos indisponivel")
	}
	return pagamentoStore.AdminPagamentos(c.UserContext(), filter)
}

func adminPagamentoFilterFromQuery(c *fiber.Ctx) (store.AdminPagamentoFilter, error) {
	status := strings.ToLower(strings.TrimSpace(c.Query("status")))
	switch status {
	case "", "pendente", "aprovado", "cancelado", "expirado", "todos":
	default:
		return store.AdminPagamentoFilter{}, errors.New("status invalido")
	}
	filter := store.AdminPagamentoFilter{Status: status}
	if inicio, err := parseAdminPagamentoStartQuery(c.Query("inicio")); err != nil {
		return store.AdminPagamentoFilter{}, errors.New("inicio invalido")
	} else {
		filter.Inicio = inicio
	}
	if fim, exclusive, err := parseAdminPagamentoEndQuery(c.Query("fim")); err != nil {
		return store.AdminPagamentoFilter{}, errors.New("fim invalido")
	} else {
		filter.Fim = fim
		filter.FimExclusive = exclusive
	}
	if filter.Inicio != nil && filter.Fim != nil && filter.Fim.Before(*filter.Inicio) {
		return store.AdminPagamentoFilter{}, errors.New("periodo invalido")
	}
	return filter, nil
}

func parseAdminPagamentoStartQuery(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if value, err := time.Parse("2006-01-02", raw); err == nil {
		return &value, nil
	}
	if value, err := time.Parse(time.RFC3339, raw); err == nil {
		utc := value.UTC()
		return &utc, nil
	}
	return nil, errors.New("data invalida")
}

func parseAdminPagamentoEndQuery(raw string) (*time.Time, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false, nil
	}
	if value, err := time.Parse("2006-01-02", raw); err == nil {
		next := value.Add(24 * time.Hour)
		return &next, true, nil
	}
	if value, err := time.Parse(time.RFC3339, raw); err == nil {
		utc := value.UTC()
		return &utc, false, nil
	}
	return nil, false, errors.New("data invalida")
}

func adminPagamentoTotals(pagamentos []store.AdminPagamento) store.AdminPixTotals {
	var totals store.AdminPixTotals
	var cents int64
	for _, pagamento := range pagamentos {
		switch pagamento.Status {
		case "pendente":
			totals.Pendente++
		case "aprovado":
			totals.Aprovado++
		case "cancelado":
			totals.Cancelado++
		case "expirado":
			totals.Expirado++
		}
		cents += parseMoneyCents(pagamento.Valor)
	}
	totals.ValorTotal = formatMoneyCents(cents)
	return totals
}

func parseMoneyCents(value string) int64 {
	normalized := strings.ReplaceAll(strings.TrimSpace(value), ",", ".")
	parts := strings.SplitN(normalized, ".", 2)
	units, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0
	}
	var decimals string
	if len(parts) == 2 {
		decimals = parts[1]
	}
	decimals += "00"
	decimalCents, err := strconv.ParseInt(decimals[:2], 10, 64)
	if err != nil {
		return 0
	}
	if units < 0 {
		return units*100 - decimalCents
	}
	return units*100 + decimalCents
}

func formatMoneyCents(cents int64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

func adminPagamentoCSVRecord(pagamento store.AdminPagamento) []string {
	return []string{
		pagamento.TXID,
		pagamento.Status,
		pagamento.Valor,
		pagamento.Descricao,
		pagamento.MAC,
		strconv.Itoa(pagamento.PlanoID),
		pagamento.Plano.Nome,
		formatCSVTime(&pagamento.CreatedAt),
		formatCSVTime(&pagamento.ExpiraEm),
	}
}
