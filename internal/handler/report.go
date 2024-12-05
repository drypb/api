package handler

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/drypb/api/internal/analysis"
	"github.com/drypb/api/internal/config"
	"github.com/gofiber/fiber/v2"
)

func GetReport(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return badRequestResponse(c, ErrEmptyID)
	}

	path := filepath.Join(config.ReportPath, id+".json")

	// Verifica se o arquivo existe
	_, err := os.Stat(path)
	if err != nil {
		return NotFoundResponse(c)
	}

	file, err := os.Open(path)
	if err != nil {
		return serverErrorResponse(c, err)
	}

	decoder := json.NewDecoder(file)
	var data analysis.Report
	err = decoder.Decode(&data)
	if err != nil {
		return serverErrorResponse(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"analysis": data,
	})
}
