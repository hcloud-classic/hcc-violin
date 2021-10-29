package config

import "github.com/Terry-Mao/goconf"

var configLocation = "/etc/hcc/violin/violin.conf"

type violinConfig struct {
	MysqlConfig     *goconf.Section
	GrpcConfig      *goconf.Section
	HTTPConfig      *goconf.Section
	RabbitMQConfig  *goconf.Section
	FluteConfig     *goconf.Section
	CelloConfig     *goconf.Section
	HarpConfig      *goconf.Section
	SchedulerConfig *goconf.Section
	PiccoloConfig   *goconf.Section
	PianoConfig     *goconf.Section
	AutoScaleConfig *goconf.Section
}

/*-----------------------------------
         Config File Example

##### CONFIG START #####
[mysql]
id user
password pass
address 111.111.111.111
port 9999
database db_name
connection_retry_count 5
connection_retry_interval_ms 500

[grpc]
port 7500
client_ping_interval_ms 1000
client_ping_timeout_ms 1000

[http]
request_timeout_ms 5000

[rabbitmq]
rabbitmq_id user
rabbitmq_password pass
rabbitmq_address 555.555.555.555
rabbitmq_port 15672

[flute]
flute_server_address 222.222.222.222
flute_server_port 3333
flute_request_timeout_ms 5000
flute_turn_off_nodes_wait_time_sec 5
flute_turn_off_nodes_retry_counts 3
flute_turn_on_nodes_retry_counts 3
flute_wait_for_leader_node_timeout_sec 30

[cello]
cello_server_address 222.222.222.222
cello_server_port 3333
cello_request_timeout_ms 5000

[harp]
harp_server_address 222.222.222.222
harp_server_port 3333
harp_request_timeout_ms 5000
harp_wait_for_leader_node_timeout_sec 30

[piano]
piano_server_address 222.222.222.222
piano_server_port 6666
piano_request_timeout_ms 30000

[autoscale]
debug on
check_server_resource_interval_ms 3000
autoscale_trigger_cpu_usage_percent 90
autoscale_trigger_memory_usage_percent 90
-----------------------------------*/
