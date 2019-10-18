package rabbitmq

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"hcc/violin/lib/logger"
	"hcc/violin/model"
)

// UpdateSubnet : Publish 'update_subnet' queues to RabbitMQ channel
func UpdateSubnet(subnet model.Subnet) error {
	qCreate, err := Channel.QueueDeclare(
		"update_subnet",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		logger.Logger.Println("update_subnet: Failed to declare a create queue")
		return err
	}

	body, _ := json.Marshal(subnet)
	err = Channel.Publish(
		"",
		qCreate.Name,
		false,
		false,
		amqp.Publishing {
			ContentType:     "text/plain",
			ContentEncoding: "utf-8",
			Body:            body,
		})
	if err != nil {
		logger.Logger.Println("update_subnet: Failed to register publisher")
		return err
	}

	return nil
}