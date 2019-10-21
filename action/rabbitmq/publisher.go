package rabbitmq

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"hcc/violin/lib/logger"
	"hcc/violin/model"
)

// GetNodes : Publish 'get_nodes' queues to RabbitMQ channel
func GetNodes(nodeNr int, serverUUID string) error {
	qCreate, err := Channel.QueueDeclare(
		"get_nodes",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		logger.Logger.Println("get_nodes: Failed to declare a create queue")
		return err
	}

	var server model.Server
	server.UUID = serverUUID
	server.NodeNr = nodeNr

	body, _ := json.Marshal(server)
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
		logger.Logger.Println("get_nodes: Failed to register publisher")
		return err
	}

	return nil
}

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
		amqp.Publishing{
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
