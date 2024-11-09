// API provides the interface for the user that wants a malware analysis.
//
// See README and [analysis] for more details.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/drypb/api/internal/config"
	"github.com/drypb/api/internal/jsonlog"
	amqp "github.com/rabbitmq/amqp091-go"
)

const version = "1.0.0"

type application struct {
	config *config.Config
	logger *jsonlog.Logger
	queue  *amqp.Connection
}

func main() {
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.PrintError(err, nil)
	}

	app := &application{
		config: cfg,
		logger: logger,
		queue:  newQueue(cfg),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      app.routes(),
		ErrorLog:     log.New(logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	app.createEssentialDirs()
	app.startWorkers()

	logger.PrintInfo("starting server", map[string]string{
		"addr":        srv.Addr,
		"env":         cfg.Env,
		"max_workers": strconv.Itoa(cfg.Queue.MaxWorkers),
	})
	err = srv.ListenAndServe()
	logger.PrintFatal(err, nil)
}
