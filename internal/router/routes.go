package router

import (
	"github.com/drypb/api/internal/handler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func SetupRoutes(app *fiber.App) {
	v1 := app.Group("v1")

	v1.Get("/healthcheck", handler.CheckHealth)

	v1.Post("/analysis", handler.StartAnalysis)

	v1.Get("/report/:id", handler.GetReport)

	v1.Get("/status/:id", websocket.New(handler.GetStatus))

	app.Use(func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotFound)
	})

}
