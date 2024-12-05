package handler

import (
	"path/filepath"
	"strconv"

	"github.com/drypb/api/internal/config"
	"github.com/drypb/api/internal/queue"
	"github.com/drypb/api/internal/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func StartAnalysis(c *fiber.Ctx) error {
	templateString := c.FormValue("template")
	file, err := c.FormFile("file")
	if err != nil {
		return badRequestResponse(c, err)
	}

	v := validator.New()
	v.Check(templateString != "", "template", "must be provided")
	v.Check(validator.PermittedValue(templateString, "105", "9011"), "template", "must be 9011")
	if !v.Valid() {
		return failedValidationResponse(c, v.Errors)
	}

	template, err := strconv.Atoi(templateString)
	if err != nil {
		return serverErrorResponse(c, err)
	}

	id := uuid.New().String()
	ext := filepath.Ext(file.Filename)
	path := filepath.Join(config.SamplePath, id+ext)
	err = c.SaveFile(file, path)
	if err != nil {
		return serverErrorResponse(c, err)
	}

	j := &queue.Job{
		ID:       id,
		File:     file,
		Template: template,
	}
	analysisQueue := queue.GetAnalysisQueue()
	analysisQueue.Enqueue(j)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": id,
	})
}
