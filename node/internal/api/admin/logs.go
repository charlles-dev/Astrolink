package admin

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"strings"
	"time"

	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

type OperationalLog = store.AdminLog

type OperationalLogFilter = store.AdminLogFilter

type operationLogStore = store.AdminLogStore

func logsHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logs, err := adminLogs(c.UserContext(), deps, operationLogFilterFromQuery(c))
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar logs")
		}
		return c.JSON(fiber.Map{
			"total": len(logs),
			"logs":  logs,
		})
	}
}

func exportLogsCSVHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logs, err := adminLogs(c.UserContext(), deps, operationLogFilterFromQuery(c))
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao exportar logs")
		}

		var buffer bytes.Buffer
		writer := csv.NewWriter(&buffer)
		if err := writer.Write([]string{"timestamp", "nivel", "tipo", "mensagem", "detalhes"}); err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao exportar logs")
		}
		for _, log := range logs {
			if err := writer.Write(operationLogCSVRecord(log)); err != nil {
				return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao exportar logs")
			}
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao exportar logs")
		}
		c.Type("csv")
		c.Set(fiber.HeaderContentDisposition, `attachment; filename="logs.csv"`)
		return c.Send(buffer.Bytes())
	}
}

func adminLogs(ctx context.Context, deps Dependencies, filter OperationalLogFilter) ([]OperationalLog, error) {
	if logStore, ok := deps.Store.(operationLogStore); ok {
		logs, err := logStore.AdminLogs(ctx, filter)
		if err != nil {
			return nil, err
		}
		if len(logs) > 0 {
			return filterOperationLogs(logs, filter), nil
		}
		allLogs, err := logStore.AdminLogs(ctx, store.AdminLogFilter{})
		if err != nil {
			return nil, err
		}
		if len(allLogs) > 0 {
			return []OperationalLog{}, nil
		}
	}
	return filterOperationLogs(localOperationLogs(time.Now().UTC()), filter), nil
}

func appendAdminLog(ctx context.Context, deps Dependencies, input store.AdminLogInput) {
	if deps.Store == nil {
		return
	}
	writer, ok := deps.Store.(store.AdminLogWriter)
	if !ok {
		return
	}
	_ = writer.AppendAdminLog(ctx, input)
}

func adminLogDetails(values map[string]any) json.RawMessage {
	if len(values) == 0 {
		return nil
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return nil
	}
	return raw
}

func operationLogFilterFromQuery(c *fiber.Ctx) OperationalLogFilter {
	return OperationalLogFilter{
		Nivel: strings.ToLower(strings.TrimSpace(c.Query("nivel"))),
		Tipo:  strings.ToLower(strings.TrimSpace(c.Query("tipo"))),
		Texto: strings.ToLower(strings.TrimSpace(c.Query("texto"))),
	}
}

func localOperationLogs(now time.Time) []OperationalLog {
	return []OperationalLog{
		{
			Timestamp: now,
			Nivel:     "info",
			Tipo:      "sistema",
			Mensagem:  "ambiente local/dev ativo sem log persistente configurado",
		},
		{
			Timestamp: now,
			Nivel:     "aviso",
			Tipo:      "backup",
			Mensagem:  "backup manual requer Postgres configurado",
		},
		{
			Timestamp: now,
			Nivel:     "info",
			Tipo:      "jobs",
			Mensagem:  "job de expiracao de sessoes disponivel para execucao operacional",
		},
	}
}

func filterOperationLogs(logs []OperationalLog, filter OperationalLogFilter) []OperationalLog {
	filtered := make([]OperationalLog, 0, len(logs))
	for _, log := range logs {
		if filter.Nivel != "" && !strings.EqualFold(log.Nivel, filter.Nivel) {
			continue
		}
		if filter.Tipo != "" && !strings.EqualFold(log.Tipo, filter.Tipo) {
			continue
		}
		if filter.Texto != "" && !operationLogContains(log, filter.Texto) {
			continue
		}
		filtered = append(filtered, log)
	}
	return filtered
}

func operationLogContains(log OperationalLog, text string) bool {
	haystack := strings.ToLower(strings.Join([]string{
		log.Nivel,
		log.Tipo,
		log.Mensagem,
		string(log.Detalhes),
	}, " "))
	return strings.Contains(haystack, text)
}

func operationLogCSVRecord(log OperationalLog) []string {
	return []string{
		log.Timestamp.UTC().Format(time.RFC3339),
		log.Nivel,
		log.Tipo,
		log.Mensagem,
		string(log.Detalhes),
	}
}
