package handler

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

var (
	ErrInvalidID = errors.New("invalid id parameter")
	ErrEmptyID   = errors.New("empty id")
)

func errorResponse(c *fiber.Ctx, status int, msg any) error {
	return c.Status(status).JSON(fiber.Map{
		"error": msg,
	})
}

func serverErrorResponse(c *fiber.Ctx, err error) error {
	return errorResponse(c, fiber.StatusInternalServerError, err.Error)
}

func NotFoundResponse(c *fiber.Ctx) error {
	msg := "the requested resource could not be found"
	return errorResponse(c, fiber.StatusNotFound, msg)
}

func methodNotAllowedResponse(c *fiber.Ctx) error {
	msg := fmt.Sprintf("the %s method is not supported for this resource", c.Method)
	return errorResponse(c, fiber.StatusMethodNotAllowed, msg)
}

func badRequestResponse(c *fiber.Ctx, err error) error {
	return errorResponse(c, fiber.StatusBadRequest, err.Error())
}

func failedValidationResponse(c *fiber.Ctx, errors map[string]string) error {
	return errorResponse(c, fiber.StatusUnprocessableEntity, fmt.Sprintf("validation failed: %v", errors))
}
