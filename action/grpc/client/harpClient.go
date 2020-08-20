package client

import (
	"context"
	"google.golang.org/grpc"
	"hcc/violin/action/grpc/pb/rpcharp"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"strconv"
	"time"
)

var harpConn *grpc.ClientConn

func initHarp() error {
	var err error

	addr := config.Harp.ServerAddress + ":" + strconv.FormatInt(config.Harp.ServerPort, 10)
	harpConn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	RC.harp = rpcharp.NewHarpClient(harpConn)
	logger.Logger.Println("gRPC harp client ready")

	return nil
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
