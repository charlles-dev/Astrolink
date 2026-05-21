package api

import (
	"log/slog"
	"time"

	"github.com/astrolink/node/internal/api/admin"
	"github.com/astrolink/node/internal/api/portal"
	"github.com/astrolink/node/internal/config"
	"github.com/astrolink/node/internal/gateway"
	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

type Dependencies struct {
	Config  config.Config
	Logger  *slog.Logger
	Store   store.Store
	Gateway gateway.Controller
}

func NewServer(deps Dependencies) *fiber.App {
	if deps.Gateway == nil {
		deps.Gateway = gateway.NoopController{}
	}
	app := fiber.New(fiber.Config{
		AppName:     "Astrolink Node",
		ReadTimeout: 10 * time.Second,
	})

	app.Get("/api/saude", func(c *fiber.Ctx) error {
		health := deps.Store.Health(c.UserContext())
		return c.JSON(fiber.Map{
			"status":          "healthy",
			"versao":          "0.1.0",
			"node":            deps.Config.NodeName,
			"uptime_segundos": 0,
			"database":        health.DatabaseStatus,
		})
	})

	portal.Register(app, portal.Dependencies{
		Store:   deps.Store,
		Gateway: deps.Gateway,
		Logger:  deps.Logger,
	})
	admin.Register(app, admin.Dependencies{
		Config:  deps.Config,
		Store:   deps.Store,
		Gateway: deps.Gateway,
	})
	return app
}
