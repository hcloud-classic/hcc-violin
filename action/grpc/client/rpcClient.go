package client

import (
	"hcc/violin/action/grpc/pb/rpcflute"
	"hcc/violin/action/grpc/pb/rpcharp"
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
	cleanHarp()
	cleanFlute()
}
