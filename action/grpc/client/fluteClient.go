package client

import (
	"context"
	errors2 "errors"
	"google.golang.org/grpc"
	"hcc/violin/action/grpc/errconv"
	"hcc/violin/action/grpc/pb/rpcflute"
	pb "hcc/violin/action/grpc/pb/rpcviolin"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"strconv"
	"time"
)

var fluteConn *grpc.ClientConn

func initFlute() error {
	var err error

	addr := config.Flute.ServerAddress + ":" + strconv.FormatInt(config.Flute.ServerPort, 10)
	fluteConn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	RC.flute = rpcflute.NewFluteClient(fluteConn)
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

	resNodePowerControl, err := rc.flute.NodePowerControl(ctx, &rpcflute.ReqNodePowerControl{
		Node:       nodes,
		PowerState: rpcflute.PowerState_ON,
	})
	if err != nil {
		return err
	}

	hccErrStack := errconv.GrpcStackToHcc(&resNodePowerControl.HccErrorStack)
	errors := *hccErrStack.ConvertReportForm()
	if len(errors) != 0 && errors[0].ErrCode != 0 {
		return errors2.New(errors[0].ErrText)
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

	var powerState = rpcflute.PowerState_OFF
	if forceOff {
		powerState = rpcflute.PowerState_FORCE_OFF
	}
	resNodePowerControl, err := rc.flute.NodePowerControl(ctx, &rpcflute.ReqNodePowerControl{
		Node:       nodes,
		PowerState: powerState,
	})
	if err != nil {
		return err
	}

	hccErrStack := errconv.GrpcStackToHcc(&resNodePowerControl.HccErrorStack)
	errors := *hccErrStack.ConvertReportForm()
	if len(errors) != 0 && errors[0].ErrCode != 0 {
		return errors2.New(errors[0].ErrText)
	}

	return nil
}

// GetNode : Get infos of the node
func (rc *RPCClient) GetNode(uuid string) (*rpcflute.Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Flute.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	node, err := rc.flute.GetNode(ctx, &rpcflute.ReqGetNode{UUID: uuid})
	if err != nil {
		return nil, err
	}

	return node.Node, nil
}

// GetNodeList : Get the list of nodes by server UUID.
func (rc *RPCClient) GetNodeList(serverUUID string) ([]rpcflute.Node, error) {
	var nodeList []rpcflute.Node

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Flute.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	pnodeList, err := rc.flute.GetNodeList(ctx, &rpcflute.ReqGetNodeList{Node: &rpcflute.Node{ServerUUID: serverUUID}})
	if err != nil {
		return nil, err
	}

	for _, pnode := range pnodeList.Node {
		nodeList = append(nodeList, rpcflute.Node{
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
func (rc *RPCClient) UpdateNode(in *rpcflute.ReqUpdateNode) (*rpcflute.Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Flute.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	node, err := rc.flute.UpdateNode(ctx, in)
	if err != nil {
		return nil, err
	}

	return node.Node, nil
}
