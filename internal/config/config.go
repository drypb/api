// Package config stores relevant application data.
package config

import (
	"flag"
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
		MaxWorkers int
		Capacity   int
	}
}

var Api Config

func (c *Config) Init() {
	flag.IntVar(&c.Port, "port", 4000, "API server port")

	flag.StringVar(&c.Env, "env", "development", "Environment (development|staging|production)")

	flag.IntVar(&c.Queue.MaxWorkers, "queueMaxWorkers", 10, "Maximum number of parallel workers")
	flag.IntVar(&c.Queue.MaxWorkers, "queueCapacity", 100, "Capacity of the queue")

	flag.Parse()
}
