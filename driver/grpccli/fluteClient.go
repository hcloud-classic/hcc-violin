package grpccli

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"hcc/violin/action/grpc/rpcflute"
	pb "hcc/violin/action/grpc/rpcviolin"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"strconv"
	"time"
)

var fluteConn *grpc.ClientConn

func initFlute() error {
	var err error

	addr := config.Flute.ServerAddress + ":" + strconv.FormatInt(config.Flute.ServerPort, 10)
	logger.Logger.Println("Trying to connect to flute module (" + addr + ")")

	for i := 0; i < int(config.Flute.ConnectionRetryCount); i++ {
		ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Flute.ConnectionTimeOutMs)*time.Millisecond)
		fluteConn, err = grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			logger.Logger.Println("Failed to connect flute module ("+addr+"): %v", err)
			logger.Logger.Println("Re-trying to connect to flute module (" +
				strconv.Itoa(i+1) + "/" + strconv.Itoa(int(config.Flute.ConnectionRetryCount)) + ")")
			continue
		}

		RC.flute = rpcflute.NewFluteClient(fluteConn)
		logger.Logger.Println("gRPC client connected to flute module")

		return nil
	}

	return errors.New("retry count exceeded to connect flute module")
}

func cleanFlute() {
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

	_, err := rc.flute.NodePowerControl(ctx, &rpcflute.ReqNodePowerControl{
		Nodes:      nodes,
		PowerState: rpcflute.ReqNodePowerControl_ON,
	})
	if err != nil {
		return err
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
