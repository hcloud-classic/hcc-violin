package config

import "github.com/Terry-Mao/goconf"

var configLocation = "/etc/hcc/violin/violin.conf"

type violinConfig struct {
	MysqlConfig     *goconf.Section
	HTTPConfig      *goconf.Section
	RabbitMQConfig  *goconf.Section
	FluteConfig     *goconf.Section
	CelloConfig     *goconf.Section
	HarpConfig      *goconf.Section
	SchedulerConfig *goconf.Section
}
