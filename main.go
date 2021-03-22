package main

import (
	"fmt"
	"hcc/violin/action/grpc/client"
	"hcc/violin/action/grpc/server"
	"hcc/violin/action/rabbitmq"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"innogrid.com/hcloud-classic/hcc_errors"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	err := logger.Init()
	if err != nil {
		hcc_errors.SetErrLogger(logger.Logger)
		hcc_errors.NewHccError(hcc_errors.ViolinInternalInitFail, "logger.Init(): "+err.Error()).Fatal()
	}
	hcc_errors.SetErrLogger(logger.Logger)

	config.Init()

	err = mysql.Init()
	if err != nil {
		hcc_errors.NewHccError(hcc_errors.ViolinInternalInitFail, "mysql.Init(): "+err.Error()).Fatal()
	}

	err = rabbitmq.Init()
	if err != nil {
		hcc_errors.NewHccError(hcc_errors.ViolinInternalInitFail, "rabbitmq.Init(): "+err.Error()).Fatal()
	}

	err = client.Init()
	if err != nil {
		hcc_errors.NewHccError(hcc_errors.ViolinInternalInitFail, "client.Init(): "+err.Error()).Fatal()
	}
}

func end() {
	client.End()
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

	server.Init()
}
