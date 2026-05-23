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
		return sendAdminPagamentosCSV(c, pagamentos, "pagamentos.csv", "erro ao exportar pagamentos")
	}
}

func pagamentosRelatorioHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filter, err := adminPagamentoFilterFromQuery(c)
		if err != nil {
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", err.Error())
		}
		pagamentos, err := adminPagamentos(c, deps, filter)
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao gerar relatorio")
		}

		formato := strings.ToLower(strings.TrimSpace(c.Query("formato", "json")))
		switch formato {
		case "", "json":
			return c.JSON(fiber.Map{
				"periodo": fiber.Map{
					"de":  formatReportTime(filter.Inicio),
					"ate": formatReportTime(filter.Fim),
				},
				"totais":     adminPagamentoTotals(pagamentos),
				"total":      len(pagamentos),
				"pagamentos": pagamentos,
			})
		case "csv":
			return sendAdminPagamentosCSV(c, pagamentos, "relatorio-pagamentos.csv", "erro ao gerar relatorio")
		case "pdf":
			return sendAdminPagamentosPDF(c, pagamentos)
		default:
			return adminError(c, fiber.StatusBadRequest, "validacao_falhou", "formato invalido")
		}
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
	if inicio, err := parseAdminPagamentoStartQuery(firstNonEmpty(c.Query("inicio"), c.Query("de"))); err != nil {
		return store.AdminPagamentoFilter{}, errors.New("inicio invalido")
	} else {
		filter.Inicio = inicio
	}
	if fim, exclusive, err := parseAdminPagamentoEndQuery(firstNonEmpty(c.Query("fim"), c.Query("ate"))); err != nil {
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
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

func sendAdminPagamentosCSV(c *fiber.Ctx, pagamentos []store.AdminPagamento, filename, errorMessage string) error {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	if err := writer.Write([]string{"txid", "status", "valor", "descricao", "mac", "plano_id", "plano", "created_at", "expira_em"}); err != nil {
		return adminError(c, fiber.StatusInternalServerError, "erro_interno", errorMessage)
	}
	for _, pagamento := range pagamentos {
		if err := writer.Write(adminPagamentoCSVRecord(pagamento)); err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", errorMessage)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return adminError(c, fiber.StatusInternalServerError, "erro_interno", errorMessage)
	}
	c.Type("csv")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+filename+`"`)
	return c.Send(buffer.Bytes())
}

func sendAdminPagamentosPDF(c *fiber.Ctx, pagamentos []store.AdminPagamento) error {
	totals := adminPagamentoTotals(pagamentos)
	lines := []string{
		"Astrolink - Relatorio de pagamentos",
		"Total de pagamentos: " + strconv.Itoa(len(pagamentos)),
		"Valor total: R$ " + strings.ReplaceAll(totals.ValorTotal, ".", ","),
		"Pendentes: " + strconv.Itoa(totals.Pendente) + " | Aprovados: " + strconv.Itoa(totals.Aprovado) + " | Cancelados: " + strconv.Itoa(totals.Cancelado) + " | Expirados: " + strconv.Itoa(totals.Expirado),
		"",
		"TXID | Status | Valor | MAC | Plano",
	}
	for i, pagamento := range pagamentos {
		if i >= 28 {
			lines = append(lines, "... relatorio truncado no PDF; use CSV para exportacao completa")
			break
		}
		lines = append(lines, strings.Join([]string{
			pagamento.TXID,
			pagamento.Status,
			"R$ " + strings.ReplaceAll(pagamento.Valor, ".", ","),
			pagamento.MAC,
			pagamento.Plano.Nome,
		}, " | "))
	}
	pdf := simplePDF(lines)
	c.Type("pdf")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="relatorio-pagamentos.pdf"`)
	return c.Send(pdf)
}

func formatReportTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func simplePDF(lines []string) []byte {
	var content bytes.Buffer
	content.WriteString("BT\n/F1 11 Tf\n50 790 Td\n")
	for _, line := range lines {
		content.WriteString("(")
		content.WriteString(escapePDFText(line))
		content.WriteString(") Tj\n0 -16 Td\n")
	}
	content.WriteString("ET\n")

	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", content.Len(), content.String()),
	}

	var pdf bytes.Buffer
	pdf.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, len(objects))
	for index, object := range objects {
		offsets = append(offsets, pdf.Len())
		fmt.Fprintf(&pdf, "%d 0 obj\n%s\nendobj\n", index+1, object)
	}
	xrefOffset := pdf.Len()
	fmt.Fprintf(&pdf, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, offset := range offsets {
		fmt.Fprintf(&pdf, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&pdf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(objects)+1, xrefOffset)
	return pdf.Bytes()
}

func escapePDFText(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "(", "\\(")
	value = strings.ReplaceAll(value, ")", "\\)")
	value = strings.ReplaceAll(value, "\r", "")
	return strings.ReplaceAll(value, "\n", " ")
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
