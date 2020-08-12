package grpccli

import (
	"hcc/violin/action/grpc/rpcflute"
	"hcc/violin/action/grpc/rpcharp"
)

// RPCClient : Struct type of gRPC clients
type RPCClient struct {
	flute rpcflute.FluteClient
	harp  rpcharp.HarpClient
}

// RC : Exported variable pointed to RPCClient
var RC = &RPCClient{}

// InitGRPCClient : Initialize clients of gRPC
func InitGRPCClient() error {
	err := initFlute()
	if err != nil {
		return err
	}

	err = initHarp()
	if err != nil {
		return err
	}

	return nil
}

// CleanGRPCClient : Close connections of gRPC clients
func CleanGRPCClient() {
	cleanFlute()
	cleanHarp()
}
