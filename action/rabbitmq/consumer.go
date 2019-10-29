package rabbitmq

import (
	"encoding/json"
	"fmt"
	"hcc/violin/dao"
	"hcc/violin/lib/logger"
	"hcc/violin/model"
	"log"
)

// ViolaToViolin : Consume Viola command
func ViolaToViolin() error {
	qCreate, err := Channel.QueueDeclare(
		"viola_to_violin",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		logger.Logger.Println("ViolaToViolin: Failed to get viola_to_violin")
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
		logger.Logger.Println("ViolaToViolin: Failed to register viola_to_violin")
		return err
	}

	go func() {
		for d := range msgsCreate {
			log.Printf("ViolaToViolin: Received a create message: %s", d.Body)

			var control model.Control
			err = json.Unmarshal(d.Body, &control)
			if err != nil {
				logger.Logger.Println("ViolaToViolin: Failed to unmarshal viola_to_violin data")
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
