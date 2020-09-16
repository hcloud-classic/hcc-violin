package rabbitmq

import (
	"encoding/json"
	"hcc/violin/lib/logger"
	"hcc/violin/model"

	"github.com/TylerBrock/colorjson"
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
	var obj map[string]interface{}
	json.Unmarshal([]byte(body), &obj)

	// Make a custom formatter with indent set
	f := colorjson.NewFormatter()
	f.Indent = 4

	// Marshall the Colorized JSON
	s, _ := f.Marshal(obj)
	// fmt.Println(string(s))
	logger.Logger.Println("doHcc Action [", string(s), "]")

	if err != nil {
		logger.Logger.Println("ViolinToViola: Failed to register publisher")
		return err
	}

	return nil
}
