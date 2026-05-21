package admin

import "github.com/gofiber/fiber/v2"

func planosHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		planos, err := deps.Store.AdminPlanos(c.UserContext())
		if err != nil {
			return adminError(c, fiber.StatusInternalServerError, "erro_interno", "erro ao carregar planos")
		}
		return c.JSON(fiber.Map{"planos": planos})
	}
}
