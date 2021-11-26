package client

import (
	"context"
	errors2 "errors"
	"hcc/violin/action/grpc/errconv"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"innogrid.com/hcloud-classic/pb"
)

var fluteConn *grpc.ClientConn

func initFlute() error {
	var err error

	addr := config.Flute.ServerAddress + ":" + strconv.FormatInt(config.Flute.ServerPort, 10)
	fluteConn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	RC.flute = pb.NewFluteClient(fluteConn)
	logger.Logger.Println("gRPC flute client ready")

	return nil
}

func closeFlute() {
	_ = fluteConn.Close()
}

// OnNode : Turn on selected node
func (rc *RPCClient) OnNode(nodeUUID string) error {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Flute.RequestTimeoutMs)*time.Millisecond)
	defer cancel()

	var nodes []*pb.Node
	node := pb.Node{
		UUID: nodeUUID,
	}
	nodes = append(nodes, &node)

	resNodePowerControl, err := rc.flute.NodePowerControl(ctx, &pb.ReqNodePowerControl{
		Node:       nodes,
		PowerState: pb.PowerState_ON,
	})
	if err != nil {
		return err
	}

	hccErrStack := errconv.GrpcStackToHcc(resNodePowerControl.HccErrorStack)
	errors := hccErrStack.ConvertReportForm()
	if errors != nil {
		stack := *errors.Stack()
		if len(stack) != 0 && stack[0].Code() != 0 {
			return errors2.New(stack[0].Text())
		}
	}

	return nil
}

// OffNode : Turn off selected node
func (rc *RPCClient) OffNode(nodeUUID string, forceOff bool) error {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Flute.RequestTimeoutMs)*time.Millisecond)
	defer cancel()

	var nodes []*pb.Node
	node := pb.Node{
		UUID: nodeUUID,
	}
	nodes = append(nodes, &node)

	powerState := pb.PowerState_OFF
	if forceOff {
		powerState = pb.PowerState_FORCE_OFF
	}
	resNodePowerControl, err := rc.flute.NodePowerControl(ctx, &pb.ReqNodePowerControl{
		Node:       nodes,
		PowerState: powerState,
	})
	if err != nil {
		return err
	}

	hccErrStack := errconv.GrpcStackToHcc(resNodePowerControl.HccErrorStack)
	errors := hccErrStack.ConvertReportForm()
	if errors != nil {
		stack := *errors.Stack()
		if len(stack) != 0 && stack[0].Code() != 0 {
			return errors2.New(stack[0].Text())
		}
	}

	return nil
}

// GetNodePowerState : Get power state of selected node
func (rc *RPCClient) GetNodePowerState(uuid string) (*pb.ResNodePowerState, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Flute.RequestTimeoutMs)*time.Millisecond)
	defer cancel()

	resNodePowerState, err := rc.flute.GetNodePowerState(ctx, &pb.ReqNodePowerState{
		UUID: uuid,
	})
	if err != nil {
		return nil, err
	}

	return resNodePowerState, nil
}

// GetNode : Get infos of the node
func (rc *RPCClient) GetNode(uuid string) (*pb.Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Flute.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resGetNode, err := rc.flute.GetNode(ctx, &pb.ReqGetNode{UUID: uuid})
	if err != nil {
		return nil, err
	}

	hccErrStack := errconv.GrpcStackToHcc(resGetNode.HccErrorStack)
	errors := hccErrStack.ConvertReportForm()
	if errors != nil {
		stack := *errors.Stack()
		if len(stack) != 0 && stack[0].Code() != 0 {
			return nil, errors2.New(stack[0].Text())
		}
	}

	return resGetNode.Node, nil
}

// GetNodeList : Get the list of nodes by server UUID.
func (rc *RPCClient) GetNodeList(serverUUID string) ([]pb.Node, error) {
	var nodeList []pb.Node

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Flute.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resGetNodeList, err := rc.flute.GetNodeList(ctx, &pb.ReqGetNodeList{Node: &pb.Node{ServerUUID: serverUUID}})
	if err != nil {
		return nil, err
	}

	hccErrStack := errconv.GrpcStackToHcc(resGetNodeList.HccErrorStack)
	errors := hccErrStack.ConvertReportForm()
	if errors != nil {
		stack := *errors.Stack()
		if len(stack) != 0 && stack[0].Code() != 0 {
			return nil, errors2.New(stack[0].Text())
		}
	}

	for _, pnode := range resGetNodeList.Node {
		nodeList = append(nodeList, pb.Node{
			UUID:        pnode.UUID,
			ServerUUID:  pnode.ServerUUID,
			BmcMacAddr:  pnode.BmcMacAddr,
			BmcIP:       pnode.BmcIP,
			PXEMacAddr:  pnode.PXEMacAddr,
			Status:      pnode.Status,
			CPUCores:    pnode.CPUCores,
			Memory:      pnode.Memory,
			Description: pnode.Description,
			Active:      pnode.Active,
			CreatedAt:   pnode.CreatedAt,
		})
	}

	return nodeList, nil
}

// UpdateNode : Update infos of the node
func (rc *RPCClient) UpdateNode(in *pb.ReqUpdateNode) (*pb.Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Flute.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resUpdateNode, err := rc.flute.UpdateNode(ctx, in)
	if err != nil {
		return nil, err
	}

	hccErrStack := errconv.GrpcStackToHcc(resUpdateNode.HccErrorStack)
	errors := hccErrStack.ConvertReportForm()
	if errors != nil {
		stack := *errors.Stack()
		if len(stack) != 0 && stack[0].Code() != 0 {
			return nil, errors2.New(stack[0].Text())
		}
	}

	return resUpdateNode.Node, nil
}
