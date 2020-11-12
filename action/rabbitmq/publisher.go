package rabbitmq

import (
	"encoding/json"
	"hcc/violin/lib/logger"
	"hcc/violin/model"

	"github.com/streadway/amqp"
)

// ViolinToViola : Publish 'run_hcc_cli' queues to RabbitMQ channel
func ViolinToViola(action model.Control) error {
	qCreate, err := Channel.QueueDeclare(
		"to_viola",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		logger.Logger.Println("ViolinToViola: Failed to declare a create queue")
		return err
	}

	body, _ := json.Marshal(action)
	err = Channel.Publish(
		"",
		qCreate.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:     "text/plain",
			ContentEncoding: "utf-8",
			Body:            body,
		})
	if err != nil {
		logger.Logger.Println("ViolinToViola: Failed to register publisher")
		return err
	}

	return nil
}
