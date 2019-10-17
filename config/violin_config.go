package config

import "github.com/Terry-Mao/goconf"

var configLocation = "/etc/violin/violin.conf"

type violinConfig struct {
	MysqlConfig *goconf.Section
	HTTPConfig  *goconf.Section
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
-----------------------------------*/
