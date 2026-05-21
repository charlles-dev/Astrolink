package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/astrolink/node/internal/api"
	"github.com/astrolink/node/internal/config"
	"github.com/astrolink/node/internal/infra/memory"
	"github.com/astrolink/node/internal/infra/postgres"
	"github.com/astrolink/node/internal/store"
)

func main() {
	cfg := config.FromEnv()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	appStore := buildStore(cfg, logger)
	app := api.NewServer(api.Dependencies{
		Config: cfg,
		Logger: logger,
		Store:  appStore,
	})

	logger.Info("starting astrolink node", "addr", cfg.HTTPAddr)
	if err := app.Listen(cfg.HTTPAddr); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func buildStore(cfg config.Config, logger *slog.Logger) store.Store {
	if cfg.DatabaseURL == "" {
		logger.Warn("DATABASE_URL ausente; usando store em memoria")
		return memory.NewStore()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	db, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Warn("Postgres indisponivel; usando store em memoria", "error", err)
		return memory.NewStore()
	}
	logger.Info("Postgres conectado; usando store persistente")
	return postgres.NewStore(db, time.Now)
}
