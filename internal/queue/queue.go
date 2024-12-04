package queue

import (
	"github.com/drypb/api/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

var Conn *amqp.Connection

func Init() error {
	var err error

	if Conn, err = amqp.Dial(config.Api.Queue.URL); err != nil {
		return err
	}
	return nil
}
