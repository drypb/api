package main

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/drypb/api/internal/data"
	"github.com/drypb/api/internal/validator"
	"github.com/google/uuid"

	amqp "github.com/rabbitmq/amqp091-go"
)

// startAnalysisHandler starts an analysis.
func (app *application) startAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	templateString := r.FormValue("template")
	file, header, err := r.FormFile("file")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	defer file.Close()

	// TODO
	v := validator.New()
	v.Check(templateString != "", "template", "must be provided")
	v.Check(validator.PermittedValue(templateString, "9011"), "template", "must be 9011")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	ext := filepath.Ext(header.Filename)
	id := uuid.New().String()
	template, err := strconv.Atoi(templateString)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Create file.
	samplePath := filepath.Join(data.DefaultSamplePath, id+ext)
	sample, err := os.Create(samplePath)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer sample.Close()

	// Copy uploaded file content to the newly created sample file.
	_, err = io.Copy(sample, file)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	ch, err := app.queue.conn.Channel()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
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
		app.serverErrorResponse(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := struct {
		Header   *multipart.FileHeader `json:"header"`
		ID       string                `json:"id"`
		Template int                   `json:"template"`
	}{
		Header:   header,
		ID:       id,
		Template: template,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = ch.PublishWithContext(
		ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/application-json",
			Body:         jsonBody,
		},
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{"id": id}
	app.writeJSON(w, http.StatusCreated, envelope{"analysis": data}, nil)
}
