package client

import (
	"hcc/violin/action/grpc/pb/rpccello"
	"hcc/violin/action/grpc/pb/rpcflute"
	"hcc/violin/action/grpc/pb/rpcharp"
	"hcc/violin/action/grpc/pb/rpcviolin_scheduler"
)

// RPCClient : Struct type of gRPC clients
type RPCClient struct {
	flute     rpcflute.FluteClient
	harp      rpcharp.HarpClient
	cello     rpccello.CelloClient
	scheduler rpcviolin_scheduler.SchedulerClient
}

// RC : Exported variable pointed to RPCClient
var RC = &RPCClient{}

// Init : Initialize clients of gRPC
func Init() error {
	err := initFlute()
	if err != nil {
		return err
	}

	err = initHarp()
	if err != nil {
		return err
	}

	err = initCello()
	if err != nil {
		return err
	}

	err = initScheduler()
	if err != nil {
		return err
	}

	return nil
}

// End : Close connections of gRPC clients
func End() {
	closeScheduler()
	closeCello()
	closeHarp()
	closeFlute()
}
