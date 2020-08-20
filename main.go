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
	"log"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	err := syscheck.CheckRoot()
	if err != nil {
		log.Fatalf("syscheck.CheckRoot(): %v", err.Error())
	}

	err = logger.Init()
	if err != nil {
		log.Fatalf("logger.Init(): %v", err.Error())
	}

	config.Parser()

	err = mysql.Init()
	if err != nil {
		logger.Logger.Fatalf("mysql.Init(): %v", err.Error())
	}

	err = rabbitmq.Init()
	if err != nil {
		logger.Logger.Fatalf("rabbitmq.Init(): %v", err.Error())
	}

	err = grpccli.InitGRPCClient()
	if err != nil {
		logger.Logger.Fatalf("grpccli.InitGRPCClient(): %v", err.Error())
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
	go func() {
		<-sigChan
		end()
		fmt.Println("Exiting violin module...")
		os.Exit(0)
	}()

	grpcsrv.Init()
}
