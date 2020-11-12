package rabbitmq

import (
	"encoding/json"
	"hcc/violin/dao"
	"hcc/violin/lib/logger"
	"hcc/violin/model"
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
			args := make(map[string]interface{})
			args["uuid"] = control.Control.HccType.ServerUUID
			args["status"] = control.Control.ActionResult // Running, Failed
			//TODO: queue get_nodes to flute module
			_, err = dao.UpdateServer(args)

			logger.Logger.Println(" UUID = " + control.Control.HccType.ServerUUID + ": " + control.Control.ActionResult)

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
