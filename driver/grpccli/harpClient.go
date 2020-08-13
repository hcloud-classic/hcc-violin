package grpccli

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"hcc/violin/action/grpc/rpcharp"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"strconv"
	"time"
)

var harpConn *grpc.ClientConn

func initHarp() error {
	var err error

	addr := config.Harp.ServerAddress + ":" + strconv.FormatInt(config.Harp.ServerPort, 10)
	logger.Logger.Println("Trying to connect to harp module (" + addr + ")")
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Harp.ConnectionTimeOutMs)*time.Millisecond)

	for i := 0; i < int(config.Harp.ConnectionRetryCount); i++ {
		harpConn, err = grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			logger.Logger.Println("Failed to connect harp module ("+addr+"): %v", err)
			continue
		}

		RC.harp = rpcharp.NewHarpClient(harpConn)
		logger.Logger.Println("gRPC client connected to harp module")

		return nil
	}

	return errors.New("retry count exceeded to connect harp module")
}

func cleanHarp() {
	_ = harpConn.Close()
}

// GetSubnet : Get infos of the subnet
func (rc *RPCClient) GetSubnet(uuid string) (*rpcharp.Subnet, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Flute.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	subnet, err := rc.harp.GetSubnet(ctx, &rpcharp.ReqGetSubnet{UUID: uuid})
	if err != nil {
		return nil, err
	}

	return subnet.Subnet, nil
}

// UpdateSubnet : Update infos of the subnet
func (rc *RPCClient) UpdateSubnet(in *rpcharp.ReqUpdateSubnet) error {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Flute.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	_, err := rc.harp.UpdateSubnet(ctx, in)
	if err != nil {
		return err
	}

	return nil
}

// CreateDHCPDConfig : Do dhcpd config file creation works
func (rc *RPCClient) CreateDHCPDConfig(subnetUUID string, nodeUUIDs string) error {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Flute.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	_, err := rc.harp.CreateDHPCDConf(ctx, &rpcharp.ReqCreateDHPCDConf{
		SubnetUUID: subnetUUID,
		NodeUUIDs:  nodeUUIDs,
	})
	if err != nil {
		return err
	}

	return nil
}
