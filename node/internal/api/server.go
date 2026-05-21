package api

import (
	"log/slog"
	"time"

	"github.com/astrolink/node/internal/api/admin"
	"github.com/astrolink/node/internal/api/portal"
	"github.com/astrolink/node/internal/config"
	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

type Dependencies struct {
	Config config.Config
	Logger *slog.Logger
	Store  store.Store
}

func NewServer(deps Dependencies) *fiber.App {
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

	portal.Register(app, deps.Store)
	admin.Register(app, admin.Dependencies{
		Config: deps.Config,
		Store:  deps.Store,
	})
	return app
}
