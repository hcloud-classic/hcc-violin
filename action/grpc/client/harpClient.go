package client

import (
	"context"
	errors2 "errors"
	"google.golang.org/grpc"
	"hcc/violin/action/grpc/errconv"
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

func closeHarp() {
	_ = harpConn.Close()
}

// GetSubnet : Get infos of the subnet
func (rc *RPCClient) GetSubnet(uuid string) (*rpcharp.Subnet, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Harp.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resGetSubnet, err := rc.harp.GetSubnet(ctx, &rpcharp.ReqGetSubnet{UUID: uuid})
	if err != nil {
		return nil, err
	}

	hccErrStack := errconv.GrpcStackToHcc(&resGetSubnet.HccErrorStack)
	errors := *hccErrStack.ConvertReportForm()
	if len(errors) != 0 && errors[0].ErrCode != 0 {
		return nil, errors2.New(errors[0].ErrText)
	}

	return resGetSubnet.Subnet, nil
}

// UpdateSubnet : Update infos of the subnet
func (rc *RPCClient) UpdateSubnet(in *rpcharp.ReqUpdateSubnet) error {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Harp.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resUpdateSubnet, err := rc.harp.UpdateSubnet(ctx, in)
	if err != nil {
		return err
	}

	hccErrStack := errconv.GrpcStackToHcc(&resUpdateSubnet.HccErrorStack)
	errors := *hccErrStack.ConvertReportForm()
	if len(errors) != 0 && errors[0].ErrCode != 0 {
		return errors2.New(errors[0].ErrText)
	}

	return nil
}

// CreateDHCPDConfig : Do dhcpd config file creation works
func (rc *RPCClient) CreateDHCPDConfig(subnetUUID string, nodeUUIDs string) error {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Harp.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resCreateDHPCDConf, err := rc.harp.CreateDHPCDConf(ctx, &rpcharp.ReqCreateDHPCDConf{
		SubnetUUID: subnetUUID,
		NodeUUIDs:  nodeUUIDs,
	})
	if err != nil {
		return err
	}

	hccErrStack := errconv.GrpcStackToHcc(&resCreateDHPCDConf.HccErrorStack)
	errors := *hccErrStack.ConvertReportForm()
	if len(errors) != 0 && errors[0].ErrCode != 0 {
		return errors2.New(errors[0].ErrText)
	}

	return nil
}
