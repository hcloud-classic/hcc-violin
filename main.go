package main

import (
	"fmt"
	"hcc/violin/action/rabbitmq"
	"hcc/violin/driver/grpccli"
	"hcc/violin/driver/grpcsrv"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"hcc/violin/lib/syscheck"
	"os"
	"os/signal"
	"syscall"
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

func end() {
	grpccli.CleanGRPCClient()
	rabbitmq.End()
	mysql.End()
	logger.End()
}

func main() {
	// Catch the exit signal
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func(){
		<- sigChan
		end()
		fmt.Println("Exiting violin module...")
		os.Exit(0)
	}()

	grpcsrv.Init()
}
