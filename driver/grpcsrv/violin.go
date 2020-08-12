package grpcsrv

import (
	"context"
	pb "hcc/violin/action/grpc/rpcviolin"
	"hcc/violin/dao"
	"hcc/violin/lib/logger"
)

type violinServer struct {
	pb.UnimplementedViolinServer
}

func returnServer(server *pb.Server) *pb.Server {
	return &pb.Server{
		UUID:       server.UUID,
		SubnetUUID: server.SubnetUUID,
		OS:         server.OS,
		ServerName: server.ServerName,
		ServerDesc: server.ServerDesc,
		CPU:        server.CPU,
		Memory:     server.Memory,
		DiskSize:   server.DiskSize,
		Status:     server.Status,
		UserUUID:   server.UserUUID,
		CreatedAt:  server.CreatedAt,
	}
}

func returnServerNode(serverNode *pb.ServerNode) *pb.ServerNode {
	return &pb.ServerNode{
		UUID:       serverNode.UUID,
		ServerUUID: serverNode.ServerUUID,
		NodeUUID:   serverNode.NodeUUID,
		CreatedAt:  serverNode.CreatedAt,
	}
}

func (s *violinServer) CreateServer(_ context.Context, in *pb.ReqCreateServer) (*pb.ResCreateServer, error) {
	logger.Logger.Println("Request received: CreateServer()")

	server, err := dao.CreateServer(in)
	if err != nil {
		return nil, err
	}

	return &pb.ResCreateServer{Server: returnServer(server)}, nil
}

func (s *violinServer) GetServer(_ context.Context, in *pb.ReqGetServer) (*pb.ResGetServer, error) {
	logger.Logger.Println("Request received: GetServer()")

	server, err := dao.ReadServer(in.GetUUID())
	if err != nil {
		return nil, err
	}

	return &pb.ResGetServer{Server: returnServer(server)}, nil
}

func (s *violinServer) GetServerList(_ context.Context, in *pb.ReqGetServerList) (*pb.ResGetServerList, error) {
	logger.Logger.Println("Request received: GetServerList()")

	serverList, err := dao.ReadServerList(in)
	if err != nil {
		return nil, err
	}

	return serverList, nil
}

func (s *violinServer) GetServerNum(_ context.Context, _ *pb.Empty) (*pb.ResGetServerNum, error) {
	logger.Logger.Println("Request received: GetServerNum()")

	serverNum, err := dao.ReadServerNum()
	if err != nil {
		return nil, err
	}

	return serverNum, nil
}

func (s *violinServer) UpdateServer(_ context.Context, in *pb.ReqUpdateServer) (*pb.ResUpdateServer, error) {
	logger.Logger.Println("Request received: UpdateServer()")

	updateServer, err := dao.UpdateServer(in)
	if err != nil {
		return nil, err
	}

	return &pb.ResUpdateServer{Server: updateServer}, nil
}

func (s *violinServer) DeleteServer(_ context.Context, in *pb.ReqDeleteServer) (*pb.ResDeleteServer, error) {
	logger.Logger.Println("Request received: DeleteServer()")

	uuid, err := dao.DeleteServer(in)
	if err != nil {
		return nil, err
	}

	return &pb.ResDeleteServer{UUID: uuid}, nil
}

func (s *violinServer) CreateServerNode(_ context.Context, in *pb.ReqCreateServerNode) (*pb.ResCreateServerNode, error) {
	logger.Logger.Println("Request received: CreateServerNode()")

	serverNode, err := dao.CreateServerNode(in)
	if err != nil {
		return nil, err
	}

	return &pb.ResCreateServerNode{ServerNode: returnServerNode(serverNode)}, nil
}

func (s *violinServer) GetServerNode(_ context.Context, in *pb.ReqGetServerNode) (*pb.ResGetServerNode, error) {
	logger.Logger.Println("Request received: GetServerNode()")

	serverNode, err := dao.ReadServerNode(in.GetUUID())
	if err != nil {
		return nil, err
	}

	return &pb.ResGetServerNode{ServerNode: returnServerNode(serverNode)}, nil
}

func (s *violinServer) GetServerNodeList(_ context.Context, in *pb.ReqGetServerNodeList) (*pb.ResGetServerNodeList, error) {
	logger.Logger.Println("Request received: GetServerNodeList()")

	serverNodeList, err := dao.ReadServerNodeList(in)
	if err != nil {
		return nil, err
	}

	return serverNodeList, nil
}

func (s *violinServer) GetServerNodeNum(_ context.Context, in *pb.ReqGetServerNodeNum) (*pb.ResGetServerNodeNum, error) {
	logger.Logger.Println("Request received: GetServerNodeNum()")

	serverNodeNum, err := dao.ReadServerNodeNum(in)
	if err != nil {
		return nil, err
	}

	return serverNodeNum, nil
}

func (s *violinServer) DeleteServerNode(_ context.Context, in *pb.ReqDeleteServerNode) (*pb.ResDeleteServerNode, error) {
	logger.Logger.Println("Request received: DeleteServerNode()")

	serverUUID, err := dao.DeleteServerNode(in)
	if err != nil {
		return nil, err
	}

	return &pb.ResDeleteServerNode{ServerUUID: serverUUID}, nil
}
