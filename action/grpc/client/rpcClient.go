package client

import (
	"github.com/hcloud-classic/pb"
)

// RPCClient : Struct type of gRPC clients
type RPCClient struct {
	flute     pb.FluteClient
	harp      pb.HarpClient
	cello     pb.CelloClient
	scheduler pb.SchedulerClient
	piccolo   pb.PiccoloClient
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

	err = initPiccolo()
	if err != nil {
		return err
	}

	return nil
}

// End : Close connections of gRPC clients
func End() {
	closePiccolo()
	closeScheduler()
	closeCello()
	closeHarp()
	closeFlute()
}
