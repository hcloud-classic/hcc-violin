package config

import "github.com/Terry-Mao/goconf"

var configLocation = "/etc/hcc/violin/violin.conf"

type violinConfig struct {
	MysqlConfig    *goconf.Section
	HTTPConfig     *goconf.Section
	RabbitMQConfig *goconf.Section
	FluteConfig    *goconf.Section
	CelloConfig    *goconf.Section
	HarpConfig     *goconf.Section
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

[http]
port 8888
RequestTimeoutMs 5000

[rabbitmq]
rabbitmq_id user
rabbitmq_password pass
rabbitmq_address 555.555.555.555
rabbitmq_port 15672

[flute]
flute_server_address 222.222.222.222
flute_server_port 3333
flute_request_timeout_ms 5000

[cello]
cello_server_address 222.222.222.222
cello_server_port 3333
cello_request_timeout_ms 5000

[harp]
harp_server_address 222.222.222.222
harp_server_port 3333
harp_request_timeout_ms 5000
harp_wait_for_leader_node_timeout_sec 30
-----------------------------------*/
