package config

import (
	"hcc/violin/lib/logger"

	"github.com/Terry-Mao/goconf"
)

var conf = goconf.New()
var config = violinConfig{}
var err error

func parseMysql() {
	config.MysqlConfig = conf.Get("mysql")
	if config.MysqlConfig == nil {
		logger.Logger.Panicln("no mysql section")
	}

	Mysql = mysql{}
	Mysql.ID, err = config.MysqlConfig.String("id")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	Mysql.Password, err = config.MysqlConfig.String("password")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	Mysql.Address, err = config.MysqlConfig.String("address")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	Mysql.Port, err = config.MysqlConfig.Int("port")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	Mysql.Database, err = config.MysqlConfig.String("database")
	if err != nil {
		logger.Logger.Panicln(err)
	}
}

func parseGrpc() {
	config.GrpcConfig = conf.Get("grpc")
	if config.GrpcConfig == nil {
		logger.Logger.Panicln("no grpc section")
	}

	Grpc.Port, err = config.GrpcConfig.Int("port")
	if err != nil {
		logger.Logger.Panicln(err)
	}
}

func parseHTTP() {
	config.HTTPConfig = conf.Get("http")
	if config.HTTPConfig == nil {
		logger.Logger.Panicln("no http section")
	}

	HTTP = http{}
	HTTP.RequestTimeoutMs, err = config.HTTPConfig.Int("request_timeout_ms")
	if err != nil {
		logger.Logger.Panicln(err)
	}
}

func parseRabbitMQ() {
	config.RabbitMQConfig = conf.Get("rabbitmq")
	if config.RabbitMQConfig == nil {
		logger.Logger.Panicln("no rabbitmq section")
	}

	RabbitMQ = rabbitmq{}
	RabbitMQ.ID, err = config.RabbitMQConfig.String("rabbitmq_id")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	RabbitMQ.Password, err = config.RabbitMQConfig.String("rabbitmq_password")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	RabbitMQ.Address, err = config.RabbitMQConfig.String("rabbitmq_address")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	RabbitMQ.Port, err = config.RabbitMQConfig.Int("rabbitmq_port")
	if err != nil {
		logger.Logger.Panicln(err)
	}
}

func parseFlute() {
	config.FluteConfig = conf.Get("flute")
	if config.FluteConfig == nil {
		logger.Logger.Panicln("no flute section")
	}

	Flute = flute{}
	Flute.ServerAddress, err = config.FluteConfig.String("flute_server_address")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	Flute.ServerPort, err = config.FluteConfig.Int("flute_server_port")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	Flute.RequestTimeoutMs, err = config.FluteConfig.Int("flute_request_timeout_ms")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	Flute.TurnOffNodesWaitTimeSec, err = config.FluteConfig.Int("flute_turn_off_nodes_wait_time_sec")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	Flute.TurnOffNodesRetryCounts, err = config.FluteConfig.Int("flute_turn_off_nodes_retry_counts")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	Flute.TurnOnNodesRetryCounts, err = config.FluteConfig.Int("flute_turn_on_nodes_retry_counts")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	Flute.WaitForLeaderNodeTimeoutSec, err = config.FluteConfig.Int("flute_wait_for_leader_node_timeout_sec")
	if err != nil {
		logger.Logger.Panicln(err)
	}
}

func parseCello() {
	config.CelloConfig = conf.Get("cello")
	if config.CelloConfig == nil {
		logger.Logger.Panicln("no cello section")
	}

	Cello = cello{}
	Cello.ServerAddress, err = config.CelloConfig.String("cello_server_address")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	Cello.ServerPort, err = config.CelloConfig.Int("cello_server_port")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	Cello.RequestTimeoutMs, err = config.CelloConfig.Int("cello_request_timeout_ms")
	if err != nil {
		logger.Logger.Panicln(err)
	}
}

func parseHarp() {
	config.HarpConfig = conf.Get("harp")
	if config.HarpConfig == nil {
		logger.Logger.Panicln("no harp section")
	}

	Harp = harp{}
	Harp.ServerAddress, err = config.HarpConfig.String("harp_server_address")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	Harp.ServerPort, err = config.HarpConfig.Int("harp_server_port")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	Harp.RequestTimeoutMs, err = config.HarpConfig.Int("harp_request_timeout_ms")
	if err != nil {
		logger.Logger.Panicln(err)
	}
}

func parseScheduler() {
	config.SchedulerConfig = conf.Get("violin_scheduler")
	if config.SchedulerConfig == nil {
		logger.Logger.Panicln("no violin_scheduler section")
	}

	ViolinScheduler = violinScheduler{}
	ViolinScheduler.ServerAddress, err = config.SchedulerConfig.String("violin_scheduler_server_address")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	ViolinScheduler.ServerPort, err = config.SchedulerConfig.Int("violin_scheduler_server_port")
	if err != nil {
		logger.Logger.Panicln(err)
	}

	ViolinScheduler.RequestTimeoutMs, err = config.SchedulerConfig.Int("violin_scheduler_request_timeout_ms")
	if err != nil {
		logger.Logger.Panicln(err)
	}
}

// Init : Parse config file and initialize config structure
func Init() {
	if err = conf.Parse(configLocation); err != nil {
		logger.Logger.Panicln(err)
	}

	parseMysql()
	parseGrpc()
	parseHTTP()
	parseRabbitMQ()
	parseFlute()
	parseCello()
	parseHarp()
	parseScheduler()
}
