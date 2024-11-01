// API provides the interface for the user that wants a malware analysis.
//
// See README and [analysis] for more details.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/drypb/api/internal/jsonlog"
	amqp "github.com/rabbitmq/amqp091-go"
)

const version = "1.0.0"

type config struct {
	port  int
	env   string
	queue struct {
		url        string
		maxWorkers int
	}
}

type queue struct {
	conn *amqp.Connection
}

type application struct {
	config config
	logger *jsonlog.Logger
	queue  *queue
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.queue.url, "queueURL", os.Getenv("QUEUE_URL"), "Queue URL")
	flag.IntVar(&cfg.queue.maxWorkers, "queueMaxWorkers", 10, "Maximum number of parallel workers")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	app := &application{
		config: cfg,
		logger: logger,
		queue:  newQueue(cfg),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		ErrorLog:     log.New(logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	app.createEssentialDirs()

	// Spawn workers.
	for range cfg.queue.maxWorkers {
		go func() {
			err := app.consume()
			if err != nil {
				app.logger.PrintError(err, nil)
				return
			}
		}()
	}

	logger.PrintInfo("starting server", map[string]string{
		"addr":        srv.Addr,
		"env":         cfg.env,
		"max_workers": strconv.Itoa(cfg.queue.maxWorkers),
	})
	err := srv.ListenAndServe()
	logger.PrintFatal(err, nil)
}
