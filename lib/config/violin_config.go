package config

import "github.com/Terry-Mao/goconf"

var configLocation = "/etc/hcc/violin/violin.conf"

type violinConfig struct {
	RsakeyConfig    *goconf.Section
	MysqlConfig     *goconf.Section
	GrpcConfig      *goconf.Section
	HTTPConfig      *goconf.Section
	RabbitMQConfig  *goconf.Section
	HornConfig      *goconf.Section
	FluteConfig     *goconf.Section
	CelloConfig     *goconf.Section
	HarpConfig      *goconf.Section
	SchedulerConfig *goconf.Section
	PiccoloConfig   *goconf.Section
	PianoConfig     *goconf.Section
	AutoScaleConfig *goconf.Section
}

