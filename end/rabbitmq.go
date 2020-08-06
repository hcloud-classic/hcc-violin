package end

import "hcc/violin/action/rabbitmq"

func rabbitmqEnd() {
	if rabbitmq.Channel != nil {
		_ = rabbitmq.Channel.Close()
	}

	if rabbitmq.Connection != nil {
		_ = rabbitmq.Connection.Close()
	}
}
