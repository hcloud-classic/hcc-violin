package main

import (
	"hcc/violin/action/graphql"
	"hcc/violin/action/rabbitmq"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"hcc/violin/lib/syscheck"
	"net/http"
	"strconv"
)

func main() {
	if !syscheck.CheckRoot() {
		return
	}

	if !logger.Prepare() {
		return
	}
	defer func() {
		_ = logger.FpLog.Close()
	}()

	config.Parser()

	err := mysql.Prepare()
	if err != nil {
		return
	}
	defer func() {
		_ = mysql.Db.Close()
	}()

	// RabbitMQ Section
	err = rabbitmq.PrepareChannel()
	if err != nil {
		logger.Logger.Panic(err)
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
		logger.Logger.Panic(err)
	}

	go func() {
		forever := make(chan bool)
		logger.Logger.Println("RabbitMQ forever channel ready.")
		<-forever
	}()

	http.Handle("/graphql", graphql.GraphqlHandler)

	err = http.ListenAndServe(":"+strconv.Itoa(int(config.HTTP.Port)), nil)
	if err != nil {
		logger.Logger.Println("Failed to prepare http server!")
	}
	logger.Logger.Println("Server is running on port " + strconv.Itoa(int(config.HTTP.Port)))
}
