package rabbitmq

import (
	"encoding/json"
	"fmt"
	"hcc/violin/dao"
	"hcc/violin/lib/logger"
	"hcc/violin/model"
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

			// CreateVolume()
		}
	}()

	return nil
}

//ConsumeViola : Consume Viola command
func ConsumeViola() error {
	qCreate, err := Channel.QueueDeclare(
		"consume_viola",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		logger.Logger.Println("ConsumeViola: Failed to get consume_viola")
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
		logger.Logger.Println("ConsumeViola: Failed to register consume_viola")
		return err
	}

	go func() {
		for d := range msgsCreate {
			log.Printf("ConsumeViola: Received a create message: %s", d.Body)

			var control model.Control
			err = json.Unmarshal(d.Body, &control)
			if err != nil {
				logger.Logger.Println("ConsumeViola: Failed to unmarshal consume_viola data")
				// return
			}
			fmt.Println("RabbitmQ : ", control)
			//To-Do******************************/
			// Violin receive cluster veryfied message, will handle message within graphql
			// update cluster status at cello DB's status
			//*************************** */
			// status, err := controlcli.HccCli(control.HccCommand, control.HccIPRange)
			// if !status && err != nil {
			// 	logger.Logger.Println("ConsumeViola: Faild execution command [", control.HccCommand, "]")
			// } else {
			// 	logger.Logger.Println("ConsumeViola: Success execution command [", control.HccCommand, "]")

			// }
			//
			args := make(map[string]interface{})
			args["uuid"] = control.ServerUUID
			args["status"] = control.HccCommand
			//TODO: queue get_nodes to flute module
			_, err = dao.UpdateServer(args)
			//logger.Logger.Println("update_subnet: UUID = " + subnet.UUID + ": " + result)
		}
	}()

	return nil
}
