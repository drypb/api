package handler

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"time"

	"github.com/drypb/api/internal/config"
	"github.com/drypb/api/internal/queue"
	"github.com/drypb/api/internal/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	amqp "github.com/rabbitmq/amqp091-go"
)

func StartAnalysis(c *fiber.Ctx) error {
	templateString := c.FormValue("template")
	file, err := c.FormFile("file")
	if err != nil {
		return badRequestResponse(c, err)
	}

	// TODO
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

	ch, err := queue.Conn.Channel()
	if err != nil {
		return serverErrorResponse(c, err)
	}

	q, err := ch.QueueDeclare(
		"task_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return serverErrorResponse(c, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := struct {
		Header   *multipart.FileHeader `json:"header"`
		ID       string                `json:"id"`
		Template int                   `json:"template"`
	}{
		Header:   file,
		ID:       id,
		Template: template,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return serverErrorResponse(c, err)
	}

	err = ch.PublishWithContext(
		ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         jsonBody,
		},
	)
	if err != nil {
		return serverErrorResponse(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": id,
	})
}
