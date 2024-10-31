package main

import (
	"context"
	"encoding/json"
	"log"
	"mime/multipart"

	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.c3sl.ufpr.br/saci/api/internal/analysis"
)

func newQueue(cfg config) *queue {
	conn, err := amqp.Dial(cfg.queue.url)
	if err != nil {
		log.Fatal(err)
	}
	queue := queue{
		conn: conn,
	}
	return &queue
}

func (app *application) consume() error {
	ch, err := app.queue.conn.Channel()
	if err != nil {
		return err
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
		return err
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return err
	}

	tasks, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	for d := range tasks {
		var task struct {
			Header   *multipart.FileHeader `json:"header"`
			ID       string                `json:"id"`
			Template int                   `json:"template"`
		}
		err := json.Unmarshal(d.Body, &task)
		if err != nil {
			return err
		}
		a, err := analysis.New(task.Header, task.ID, task.Template)
		if err != nil {
			return err
		}
		err = a.Run(context.Background())
		if err != nil {
			a.Report.Request.Error = err.Error()
			a.Report.Save()
			a.Report.SaveAll()
			err = a.Cleanup()
			if err != nil {
				return err
			}
			return err
		}
	}

	return nil
}
