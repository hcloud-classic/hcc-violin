package rabbitmq

import (
	"hcc/violin/lib/logger"
	"log"
)

// ReturnNodes : Consume 'return_nodes' queues from RabbitMQ channel
func ReturnNodes() error {
	qCreate, err := Channel.QueueDeclare(
		"return_nodes",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		logger.Logger.Println("return_nodes: Failed to declare a create queue")
		return err
	}

	msgsCreate, err := Channel.Consume(
		qCreate.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Logger.Println("return_nodes: Failed to register consumer")
		return err
	}

	go func() {
		for d := range msgsCreate {
			log.Printf("return_nodes: Received a create message: %s", d.Body)

			//CreateVolume()
		}
	}()

	return nil
}
