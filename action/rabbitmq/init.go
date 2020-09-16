package rabbitmq

import (
	"hcc/violin/lib/logger"
)

// Init : Initialize Rabbitmq
func Init() error {
	err := prepareChannel()
	if err != nil {
		return err
	}

	err = violaToViolin()
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

// End : Close Rabbitmq channel and connection
func End() {
	if Channel != nil {
		_ = Channel.Close()
	}

	if Connection != nil {
		_ = Connection.Close()
	}
}
