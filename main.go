package main

import (
	"hcc/violin/action/graphql"
	"hcc/violin/action/rabbitmq"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"hcc/violin/lib/syscheck"
	"net/http"
	"runtime"
	"strconv"
)

func main() {
	// Use max CPUs
	runtime.GOMAXPROCS(runtime.NumCPU())

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

	//RabbitMQ Section
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
	err = rabbitmq.ConsumeViola()
	if err != nil {
		logger.Logger.Panic(err)
	}

	//var listNodeData graphql.ListNodeData
	//listNodeData, err = graphql.GetNodes()
	//if err != nil {
	//	logger.Logger.Panic(err)
	//}
	//fmt.Println(listNodeData.Data.ListNode[0].UUID)

	http.Handle("/graphql", graphql.GraphqlHandler)

	logger.Logger.Println("Server is running on port " + strconv.Itoa(int(config.HTTP.Port)))
	err = http.ListenAndServe(":"+strconv.Itoa(int(config.HTTP.Port)), nil)
	if err != nil {
		logger.Logger.Println("Failed to prepare http server!")
	}
}
