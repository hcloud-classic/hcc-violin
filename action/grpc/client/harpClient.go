package client

import (
	"context"
	errors2 "errors"
	"google.golang.org/grpc"
	"hcc/violin/action/grpc/errconv"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"innogrid.com/hcloud-classic/pb"
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

	RC.harp = pb.NewHarpClient(harpConn)
	logger.Logger.Println("gRPC harp client ready")

	return nil
}

func closeHarp() {
	_ = harpConn.Close()
}

// GetSubnet : Get infos of the subnet
func (rc *RPCClient) GetSubnet(uuid string) (*pb.Subnet, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Harp.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resGetSubnet, err := rc.harp.GetSubnet(ctx, &pb.ReqGetSubnet{UUID: uuid})
	if err != nil {
		return nil, err
	}

	hccErrStack := errconv.GrpcStackToHcc(resGetSubnet.HccErrorStack)
	errors := hccErrStack.ConvertReportForm()
	if errors != nil {
		stack := *errors.Stack()
		if len(stack) != 0 && stack[0].Code() != 0 {
			return nil, errors2.New(stack[0].Text())
		}
	}

	return resGetSubnet.Subnet, nil
}

// GetSubnetByServer : Get infos of the subnet by server UUID
func (rc *RPCClient) GetSubnetByServer(serverUUID string) (*pb.Subnet, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Harp.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resGetSubnetByServer, err := rc.harp.GetSubnetByServer(ctx, &pb.ReqGetSubnetByServer{ServerUUID: serverUUID})
	if err != nil {
		return nil, err
	}

	hccErrStack := errconv.GrpcStackToHcc(resGetSubnetByServer.HccErrorStack)
	errors := hccErrStack.ConvertReportForm()
	if errors != nil {
		stack := *errors.Stack()
		if len(stack) != 0 && stack[0].Code() != 0 {
			return nil, errors2.New(stack[0].Text())
		}
	}

	return resGetSubnetByServer.Subnet, nil
}

// UpdateSubnet : Update infos of the subnet
func (rc *RPCClient) UpdateSubnet(in *pb.ReqUpdateSubnet) error {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Harp.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resUpdateSubnet, err := rc.harp.UpdateSubnet(ctx, in)
	if err != nil {
		return err
	}

	hccErrStack := errconv.GrpcStackToHcc(resUpdateSubnet.HccErrorStack)
	errors := hccErrStack.ConvertReportForm()
	if errors != nil {
		stack := *errors.Stack()
		if len(stack) != 0 && stack[0].Code() != 0 {
			return errors2.New(stack[0].Text())
		}
	}

	return nil
}

// CreateDHCPDConfig : Do dhcpd config file creation works
func (rc *RPCClient) CreateDHCPDConfig(subnetUUID string) error {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Harp.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resCreateDHCPDConf, err := rc.harp.CreateDHCPDConf(ctx, &pb.ReqCreateDHCPDConf{
		SubnetUUID: subnetUUID,
	})
	if err != nil {
		return err
	}

	hccErrStack := errconv.GrpcStackToHcc(resCreateDHCPDConf.HccErrorStack)
	errors := hccErrStack.ConvertReportForm()
	if errors != nil {
		stack := *errors.Stack()
		if len(stack) != 0 && stack[0].Code() != 0 {
			return errors2.New(stack[0].Text())
		}
	}

	return nil
}

// DeleteDHCPDConfig : Do dhcpd config file deletion works
func (rc *RPCClient) DeleteDHCPDConfig(subnetUUID string) error {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Harp.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resDeleteDHCPDConf, err := rc.harp.DeleteDHCPDConf(ctx, &pb.ReqDeleteDHCPDConf{
		SubnetUUID: subnetUUID,
	})
	if err != nil {
		return err
	}

	hccErrStack := errconv.GrpcStackToHcc(resDeleteDHCPDConf.HccErrorStack)
	errors := hccErrStack.ConvertReportForm()
	if errors != nil {
		stack := *errors.Stack()
		if len(stack) != 0 && stack[0].Code() != 0 {
			return errors2.New(stack[0].Text())
		}
	}

	return nil
}

// DeleteAdaptiveIPServer : Delete of the adaptiveIP server
func (rc *RPCClient) DeleteAdaptiveIPServer(serverUUID string) (*pb.ResDeleteAdaptiveIPServer, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Harp.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resDeleteAdaptiveIPServer, err := rc.harp.DeleteAdaptiveIPServer(ctx, &pb.ReqDeleteAdaptiveIPServer{ServerUUID: serverUUID})
	if err != nil {
		return nil, err
	}

	hccErrStack := errconv.GrpcStackToHcc(resDeleteAdaptiveIPServer.HccErrorStack)
	errors := hccErrStack.ConvertReportForm()
	if errors != nil {
		stack := *errors.Stack()
		if len(stack) != 0 && stack[0].Code() != 0 {
			return nil, errors2.New(stack[0].Text())
		}
	}

	return resDeleteAdaptiveIPServer, nil
}
