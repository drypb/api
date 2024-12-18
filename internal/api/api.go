// Package api provides the interface for the user that wants a malware analysis.
//
// See README and [analysis] for more details.
package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/drypb/api/internal/config"
	"github.com/drypb/api/internal/queue"
	"github.com/drypb/api/internal/router"
	"github.com/gofiber/fiber/v2"
)

func Run() error {
	apiConfig := config.GetApiConfig()
	apiConfig.Init()

	analysisQueue := queue.GetAnalysisQueue()
	analysisQueue.Init()

	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  1 * time.Minute,
		JSONEncoder: func(v any) ([]byte, error) {
			return json.MarshalIndent(v, "", "  ")
		},
	})

	router.SetupRoutes(app)

	if err := createEssentialDirs(); err != nil {
		return err
	}

	addr := fmt.Sprintf(":%d", config.Api.Port)
	return app.Listen(addr)
}
