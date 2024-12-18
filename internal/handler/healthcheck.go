package handler

import (
	"github.com/drypb/api/internal/config"
	"github.com/gofiber/fiber/v2"
)

const version = "1.0.0"

func CheckHealth(c *fiber.Ctx) error {
	apiConfig := config.GetApiConfig()
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "available",
		"system_info": map[string]string{
			"environment": apiConfig.Env,
			"version":     version,
		},
	})
}
