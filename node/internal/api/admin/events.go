package admin

import (
	"bufio"
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

type eventSnapshot struct {
	Timestamp           string `json:"timestamp"`
	Database            string `json:"database"`
	UsuariosTotal       int    `json:"usuarios_total"`
	UsuariosAtivos      int    `json:"usuarios_ativos"`
	VouchersTotal       int    `json:"vouchers_total"`
	VouchersAtivos      int    `json:"vouchers_ativos"`
	PagamentosTotal     int    `json:"pagamentos_total"`
	PagamentosPendentes int    `json:"pagamentos_pendentes"`
	PagamentosAprovados int    `json:"pagamentos_aprovados"`
	LogsTotal           int    `json:"logs_total"`
}

func eventsHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		once := c.Query("once") == "1"

		c.Set(fiber.HeaderContentType, "text/event-stream")
		c.Set(fiber.HeaderCacheControl, "no-cache")
		c.Set(fiber.HeaderConnection, "keep-alive")
		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			if once {
				_ = writeEventSnapshot(w, adminEventSnapshot(ctx, deps))
				return
			}

			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if err := writeEventSnapshot(w, adminEventSnapshot(ctx, deps)); err != nil {
					return
				}

				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
				}
			}
		})
		return nil
	}
}

func adminEventSnapshot(ctx context.Context, deps Dependencies) eventSnapshot {
	snapshot := eventSnapshot{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Database:  "error",
	}
	if deps.Store == nil {
		return snapshot
	}

	health := deps.Store.Health(ctx)
	snapshot.Database = health.DatabaseStatus
	if snapshot.Database == "" {
		snapshot.Database = "ok"
	}

	if usuarios, err := deps.Store.Usuarios(ctx); err == nil {
		snapshot.UsuariosTotal = len(usuarios)
		for _, usuario := range usuarios {
			if strings.EqualFold(usuario.Status, "ativo") {
				snapshot.UsuariosAtivos++
			}
		}
	} else {
		snapshot.Database = "error"
	}

	if vouchers, err := deps.Store.AdminVouchers(ctx); err == nil {
		snapshot.VouchersTotal = len(vouchers)
		for _, voucher := range vouchers {
			if voucher.Ativo {
				snapshot.VouchersAtivos++
			}
		}
	} else {
		snapshot.Database = "error"
	}

	if pagamentoStore, ok := deps.Store.(store.AdminPagamentoStore); ok {
		if pagamentos, err := pagamentoStore.AdminPagamentos(ctx, store.AdminPagamentoFilter{Status: "todos"}); err == nil {
			snapshot.PagamentosTotal = len(pagamentos)
			for _, pagamento := range pagamentos {
				switch strings.ToLower(strings.TrimSpace(pagamento.Status)) {
				case "pendente":
					snapshot.PagamentosPendentes++
				case "aprovado":
					snapshot.PagamentosAprovados++
				}
			}
		} else {
			snapshot.Database = "error"
		}
	}

	if logs, err := adminLogs(ctx, deps, OperationalLogFilter{}); err == nil {
		snapshot.LogsTotal = len(logs)
	} else {
		snapshot.Database = "error"
	}

	return snapshot
}

func writeEventSnapshot(w *bufio.Writer, snapshot eventSnapshot) error {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	if _, err := w.WriteString("event: snapshot\n"); err != nil {
		return err
	}
	if _, err := w.WriteString("data: "); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	if _, err := w.WriteString("\n\n"); err != nil {
		return err
	}
	return w.Flush()
}
