package rabbitmq

import (
	"encoding/json"
	"hcc/violin/lib/logger"
	"hcc/violin/model"

	"github.com/streadway/amqp"
)

// RunHccCLI : Publish 'run_hcc_cli' queues to RabbitMQ channel
func RunHccCLI(action []model.Control) error {
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

type nodesInfo struct {
	ServerUUID string `json:"server_uuid"`
	NodeNr     int    `json:"node_nr"`
}

// TODO: oboe 로 부터 온 create_server 요청에 대응 필요
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

	var nodesinfo nodesInfo
	nodesinfo.ServerUUID = serverUUID
	nodesinfo.NodeNr = nodeNr

	body, _ := json.Marshal(nodesinfo)
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

// CreateVolume : Publish 'create_volume' queues to RabbitMQ channel
func CreateVolume(volume model.Volume) error {
	qCreate, err := Channel.QueueDeclare(
		"create_volume",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		logger.Logger.Println("create_volume: Failed to declare a create queue")
		return err
	}

	body, _ := json.Marshal(volume)
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
		logger.Logger.Println("create_volume: Failed to register publisher")
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
