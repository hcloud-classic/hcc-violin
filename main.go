package main

import (
	"hcc/violin/action/rabbitmq"
	"hcc/violin/driver/grpccli"
	"hcc/violin/driver/grpcsrv"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"hcc/violin/lib/syscheck"
)

func init() {
	err := syscheck.CheckRoot()
	if err != nil {
		panic(err)
	}

	err = logger.Init()
	if err != nil {
		panic(err)
	}

	config.Parser()

	err = mysql.Init()
	if err != nil {
		panic(err)
	}

	err = rabbitmq.Init()
	if err != nil {
		panic(err)
	}

	err = grpccli.InitGRPCClient()
	if err != nil {
		panic(err)
	}
}

func main() {
	defer func() {
		grpccli.CleanGRPCClient()
		rabbitmq.End()
		mysql.End()
		logger.End()
	}()

	grpcsrv.Init()
}
