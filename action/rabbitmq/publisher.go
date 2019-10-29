package rabbitmq

import (
	"encoding/json"
	"hcc/violin/lib/logger"
	"hcc/violin/model"

	"github.com/streadway/amqp"
)

// RunHccCLI : Publish 'run_hcc_cli' queues to RabbitMQ channel
func RunHccCLI(action model.Control) error {
	qCreate, err := Channel.QueueDeclare(
		"run_hcc_cli",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		logger.Logger.Println("run_hcc_cli: Failed to declare a create queue")
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
		logger.Logger.Println("get_nodes: Failed to register publisher")
		return err
	}

	return nil
}
