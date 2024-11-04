// Package config stores relevant application data.
package config

import (
	"flag"
	"fmt"
	"os"
)

const (
	ReportPath = "reports" // ReportPath is where the reports will be located.
	SamplePath = "samples" // SamplePath is where the malware samples will be located.
	LogPath    = "logs"    // LogPath is where the driver logs will be located.
	StatusPath = "status"  // StatusPath is where the status for the websocket route will be located.
)

type Config struct {
	Port  int
	Env   string
	Queue struct {
		URL        string
		MaxWorkers int
	}
}

func LoadConfig() (*Config, error) {
	var cfg Config

	flag.IntVar(&cfg.Port, "port", 4000, "API server port")

	flag.StringVar(&cfg.Queue.URL, "queueURL", os.Getenv("QUEUE_URL"), "Queue URL")
	flag.IntVar(&cfg.Queue.MaxWorkers, "queueMaxWorkers", 10, "Maximum number of parallel workers")

	flag.StringVar(&cfg.Env, "env", "development", "Environment (development|staging|production)")

	flag.Parse()

	if cfg.Queue.URL == "" {
		return nil, fmt.Errorf("Queue URL is empty")
	}

	return &cfg, nil
}
