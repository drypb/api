package queue

import (
	"context"
	"encoding/json"
	"log"
	"mime/multipart"

	"github.com/drypb/api/internal/analysis"
	"github.com/drypb/api/internal/config"
)

func StartWorkers() {
	errCh := make(chan error)

	for range config.Api.Queue.MaxWorkers {
		go func() {
			for {
				if err := consume(); err != nil {
					errCh <- err
				}
			}
		}()
	}

	go func() {
		for err := range errCh {
			log.Printf("Worker failed: %v", err)
		}
	}()
}

func consume() error {
	ch, err := Conn.Channel()
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
		if err := json.Unmarshal(d.Body, &task); err != nil {
			return err
		}
		a, err := analysis.New(task.Header, task.ID, task.Template)
		if err != nil {
			return err
		}
		if err := a.Run(context.Background()); err != nil {
			a.Report.Request.Error = err.Error()
			a.Report.Save("status")
			a.Report.Save("report")
			if err := a.Cleanup(); err != nil {
				return err
			}
			return err
		}
	}

	return nil
}
