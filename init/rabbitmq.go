package init

import (
	"hcc/violin/action/rabbitmq"
	"hcc/violin/lib/logger"
)

func rabbitmqInit() error {
	err := rabbitmq.PrepareChannel()
	if err != nil {
		return err
	}
	defer func() {
		_ = rabbitmq.Channel.Close()
	}()
	defer func() {
		_ = rabbitmq.Connection.Close()
	}()

	// Viola Section
	err = rabbitmq.ViolaToViolin()
	if err != nil {
		return err
	}

	go func() {
		forever := make(chan bool)
		logger.Logger.Println("RabbitMQ forever channel ready.")
		<-forever
	}()

	return nil
}
