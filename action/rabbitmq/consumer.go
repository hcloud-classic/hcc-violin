package rabbitmq

import (
	"encoding/json"
	"hcc/violin/action/grpc/client"
	"hcc/violin/lib/logger"
	"hcc/violin/model"
)

func violaToViolin() error {
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
			logger.Logger.Printf("ViolaToViolin: Received a create message: %s\n", d.Body)

			var control model.Control
			err = json.Unmarshal(d.Body, &control)
			if err != nil {
				logger.Logger.Println("ViolaToViolin: Failed to unmarshal viola_to_violin data")
				// return
			}
			logger.Logger.Println("[Violin]RabbitmQ Receive: ", control)
			//To-Do******************************/
			// Violin receive cluster veryfied message, will handle message within graphql
			// update cluster status at cello DB's status
			//*************************** */
			//
			uuid := control.Control.HccType.ServerUUID
			status := control.Control.ActionResult // Running, Failed
			// TODO: queue get_nodes to flute module
			err = updateServerStatus(uuid, status)
			if err != nil {
				logger.Logger.Println("ViolaToViolin: " + err.Error())
			}

			logger.Logger.Println("ViolaToViolin: UUID = " + control.Control.HccType.ServerUUID + ": " + control.Control.ActionResult)
			printLogDoCreateServerRoutineQueue(uuid, "Server Created")
			_ = client.RC.WriteServerAction(
				uuid,
				"violin / server_running ",
				"Success",
				"",
				control.Token)
			// vntOpt := model.Vnc{
			// 	ServerUUID=args["uuid"].(string)
			// }
			// Leaderiprange :=  strings.Split(control.Control.HccType.HccIPRange," ")
			// if len(Leaderiprange)>1{
			// 	vntOpt.TargetIP:=Leaderiprange[0]
			// 	CreateVnc, actionerr := driver.VncControl()
			// }

		}
	}()

	return nil
}

// ConsumeCreateServer : Consume server creating queues from RabbitMQ channel
func ConsumeCreateServer() error {
	qCreate, err := Channel.QueueDeclare(
		"create_server",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		logger.Logger.Println("QueueCreateServer: Failed to get create_server")
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
		logger.Logger.Println("QueueCreateServer: Failed to register create_server")
		return err
	}

	go func() {
		for d := range msgsCreate {
			logger.Logger.Printf("QueueCreateServer: Received a message: %s\n", d.Body)

			var data createServerDataStruct
			err = json.Unmarshal(d.Body, &data)
			if err != nil {
				logger.Logger.Println("QueueCreateServer: Failed to unmarshal create_server data")
				return
			}

			switch data.Action {
			case "create":
				logger.Logger.Println("QueueCreateServer: Creating server for " + data.RoutineServerUUID)
				DoCreateServerRoutineQueue(data.RoutineServerUUID, data.RoutineServerOS, &data.RoutineSubnet, data.RoutineNodes,
					data.CelloParams, data.RoutineFirstIP, data.RoutineLastIP, data.Token)

				break
			case "update":
				logger.Logger.Println("QueueUpdateServerNodes: Updating server for " + data.RoutineServerUUID)
				DoUpdateServerNodesRoutineQueue(data.RoutineServerUUID, &data.RoutineSubnet, data.RoutineNodes,
					data.RoutineFirstIP, data.RoutineLastIP, data.Token)

				break
			case "delete":
				logger.Logger.Println("QueueUDeleteServerNodes: Deleting server for " + data.RoutineServerUUID)
				DoDeleteServerRoutineQueue(data.RoutineServerUUID, data.Token)

				break
			}
		}
	}()

	return nil
}
