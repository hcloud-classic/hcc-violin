package client

import (
	"innogrid.com/hcloud-classic/pb"
)

// RPCClient : Struct type of gRPC clients
type RPCClient struct {
	horn      pb.HornClient
	flute     pb.FluteClient
	harp      pb.HarpClient
	cello     pb.CelloClient
	scheduler pb.SchedulerClient
	piccolo   pb.PiccoloClient
	piano     pb.PianoClient
}

// RC : Exported variable pointed to RPCClient
var RC = &RPCClient{}

// Init : Initialize clients of gRPC
func Init() error {
	err := initHorn()
	if err != nil {
		return err
	}

	err = initFlute()
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

	err = initPiano()
	if err != nil {
		return err
	}
	checkPiano()

	return nil
}

// End : Close connections of gRPC clients
func End() {
	closePiano()
	closePiccolo()
	closeScheduler()
	closeCello()
	closeHarp()
	closeFlute()
	closeHorn()
}
