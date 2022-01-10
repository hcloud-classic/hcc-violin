package main

import (
	"fmt"
	"hcc/violin/action/grpc/client"
	"hcc/violin/action/grpc/server"
	"hcc/violin/action/rabbitmq"
	"hcc/violin/lib/autoscale"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"hcc/violin/lib/pid"
	"innogrid.com/hcloud-classic/hcc_errors"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func init() {
	violinRunning, violinPID, err := pid.IsViolinRunning()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	if violinRunning {
		fmt.Println("violin is already running. (PID: " + strconv.Itoa(violinPID) + ")")
		os.Exit(1)
	}
	err = pid.WriteViolinPID()
	if err != nil {
		_ = pid.DeleteViolinPID()
		fmt.Println(err)
		panic(err)
	}

	err = logger.Init()
	if err != nil {
		hcc_errors.SetErrLogger(logger.Logger)
		hcc_errors.NewHccError(hcc_errors.ViolinInternalInitFail, "logger.Init(): "+err.Error()).Fatal()
		_ = pid.DeleteViolinPID()
	}
	hcc_errors.SetErrLogger(logger.Logger)

	config.Init()

	err = client.Init()
	if err != nil {
		hcc_errors.NewHccError(hcc_errors.ViolinInternalInitFail, "client.Init(): "+err.Error()).Fatal()
		_ = pid.DeleteViolinPID()
	}

	err = mysql.Init()
	if err != nil {
		hcc_errors.NewHccError(hcc_errors.ViolinInternalInitFail, "mysql.Init(): "+err.Error()).Fatal()
		_ = pid.DeleteViolinPID()
	}

	err = rabbitmq.Init()
	if err != nil {
		hcc_errors.NewHccError(hcc_errors.ViolinInternalInitFail, "rabbitmq.Init(): "+err.Error()).Fatal()
		_ = pid.DeleteViolinPID()
	}

	logger.Logger.Println("Starting autoscale.CheckServerResource() Interval is " + strconv.Itoa(int(config.AutoScale.CheckServerResourceIntervalMs)) + "ms")
	autoscale.CheckServerResource()
}

func end() {
	rabbitmq.End()
	mysql.End()
	client.End()
	logger.End()
	_ = pid.DeleteViolinPID()
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
