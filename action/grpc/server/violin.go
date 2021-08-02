package server

import (
	"context"
	"hcc/violin/action/grpc/errconv"
	"hcc/violin/dao"
	"hcc/violin/daoext"
	"hcc/violin/lib/logger"
	"innogrid.com/hcloud-classic/hcc_errors"
	"innogrid.com/hcloud-classic/pb"
)

type violinServer struct {
	pb.UnimplementedViolinServer
}

func returnServer(server *pb.Server) *pb.Server {
	return &pb.Server{
		UUID:       server.UUID,
		GroupID:    server.GroupID,
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

	server, errStack := dao.CreateServer(in)
	if server == nil {
		return &pb.ResCreateServer{Server: &pb.Server{}, HccErrorStack: errconv.HccStackToGrpc(errStack)}, nil
	}

	return &pb.ResCreateServer{Server: returnServer(server)}, nil
}

func (s *violinServer) GetServer(_ context.Context, in *pb.ReqGetServer) (*pb.ResGetServer, error) {
	logger.Logger.Println("Request received: GetServer()")

	server, errCode, errStr := dao.ReadServer(in.GetUUID())
	if errCode != 0 {
		errStack := hcc_errors.NewHccErrorStack(hcc_errors.NewHccError(errCode, errStr))
		return &pb.ResGetServer{Server: &pb.Server{}, HccErrorStack: errconv.HccStackToGrpc(errStack)}, nil
	}

	return &pb.ResGetServer{Server: returnServer(server)}, nil
}

func (s *violinServer) GetServerList(_ context.Context, in *pb.ReqGetServerList) (*pb.ResGetServerList, error) {
	// logger.Logger.Println("Request received: GetServerList()")

	serverList, errCode, errStr := dao.ReadServerList(in)
	if errCode != 0 {
		errStack := hcc_errors.NewHccErrorStack(hcc_errors.NewHccError(errCode, errStr))
		return &pb.ResGetServerList{Server: []*pb.Server{}, HccErrorStack: errconv.HccStackToGrpc(errStack)}, nil
	}

	return serverList, nil
}

func (s *violinServer) GetServerNum(_ context.Context, in *pb.ReqGetServerNum) (*pb.ResGetServerNum, error) {
	// logger.Logger.Println("Request received: GetServerNum()")

	serverNum, errCode, errStr := dao.ReadServerNum(in)
	if errCode != 0 {
		errStack := hcc_errors.NewHccErrorStack(hcc_errors.NewHccError(errCode, errStr))
		return &pb.ResGetServerNum{Num: 0, HccErrorStack: errconv.HccStackToGrpc(errStack)}, nil
	}

	return serverNum, nil
}

func (s *violinServer) UpdateServer(_ context.Context, in *pb.ReqUpdateServer) (*pb.ResUpdateServer, error) {
	logger.Logger.Println("Request received: UpdateServer()")

	updateServer, errStack := dao.UpdateServer(in)
	if updateServer == nil {
		return &pb.ResUpdateServer{Server: &pb.Server{}, HccErrorStack: errconv.HccStackToGrpc(errStack)}, nil
	}

	return &pb.ResUpdateServer{Server: updateServer}, nil
}

func (s *violinServer) UpdateServerNodes(_ context.Context, in *pb.ReqUpdateServerNodes) (*pb.ResUpdateServerNodes, error) {
	logger.Logger.Println("Request received: UpdateServer()")

	updateServer, errStack := dao.UpdateServerNodes(in)
	if updateServer == nil {
		return &pb.ResUpdateServerNodes{Server: &pb.Server{}, HccErrorStack: errconv.HccStackToGrpc(errStack)}, nil
	}

	return &pb.ResUpdateServerNodes{Server: updateServer}, nil
}

func (s *violinServer) DeleteServer(_ context.Context, in *pb.ReqDeleteServer) (*pb.ResDeleteServer, error) {
	logger.Logger.Println("Request received: DeleteServer()")

	deleteServer, errCode, errStr := dao.DeleteServer(in)
	if errCode != 0 {
		errStack := hcc_errors.NewHccErrorStack(hcc_errors.NewHccError(errCode, errStr))
		return &pb.ResDeleteServer{Server: &pb.Server{}, HccErrorStack: errconv.HccStackToGrpc(errStack)}, nil
	}

	return &pb.ResDeleteServer{Server: deleteServer}, nil
}

func (s *violinServer) CreateServerNode(_ context.Context, in *pb.ReqCreateServerNode) (*pb.ResCreateServerNode, error) {
	logger.Logger.Println("Request received: CreateServerNode()")

	serverNode, errCode, errStr := dao.CreateServerNode(in)
	if errCode != 0 {
		errStack := hcc_errors.NewHccErrorStack(hcc_errors.NewHccError(errCode, errStr))
		return &pb.ResCreateServerNode{ServerNode: &pb.ServerNode{}, HccErrorStack: errconv.HccStackToGrpc(errStack)}, nil
	}

	return &pb.ResCreateServerNode{ServerNode: returnServerNode(serverNode)}, nil
}

func (s *violinServer) GetServerNode(_ context.Context, in *pb.ReqGetServerNode) (*pb.ResGetServerNode, error) {
	logger.Logger.Println("Request received: GetServerNode()")

	serverNode, errCode, errStr := dao.ReadServerNode(in.GetUUID())
	if errCode != 0 {
		errStack := hcc_errors.NewHccErrorStack(hcc_errors.NewHccError(errCode, errStr))
		return &pb.ResGetServerNode{ServerNode: &pb.ServerNode{}, HccErrorStack: errconv.HccStackToGrpc(errStack)}, nil
	}

	return &pb.ResGetServerNode{ServerNode: returnServerNode(serverNode)}, nil
}

func (s *violinServer) GetServerNodeList(_ context.Context, in *pb.ReqGetServerNodeList) (*pb.ResGetServerNodeList, error) {
	logger.Logger.Println("Request received: GetServerNodeList()")

	serverNodeList, errCode, errStr := daoext.ReadServerNodeList(in)
	if errCode != 0 {
		errStack := hcc_errors.NewHccErrorStack(hcc_errors.NewHccError(errCode, errStr))
		return &pb.ResGetServerNodeList{ServerNode: []*pb.ServerNode{}, HccErrorStack: errconv.HccStackToGrpc(errStack)}, nil
	}

	return serverNodeList, nil
}

func (s *violinServer) GetServerNodeNum(_ context.Context, in *pb.ReqGetServerNodeNum) (*pb.ResGetServerNodeNum, error) {
	// logger.Logger.Println("Request received: GetServerNodeNum()")

	serverNodeNum, errCode, errStr := dao.ReadServerNodeNum(in)
	if errCode != 0 {
		errStack := hcc_errors.NewHccErrorStack(hcc_errors.NewHccError(errCode, errStr))
		return &pb.ResGetServerNodeNum{Num: 0, HccErrorStack: errconv.HccStackToGrpc(errStack)}, nil
	}

	return serverNodeNum, nil
}

func (s *violinServer) DeleteServerNode(_ context.Context, in *pb.ReqDeleteServerNode) (*pb.ResDeleteServerNode, error) {
	logger.Logger.Println("Request received: DeleteServerNode()")

	deleteServerNode, errCode, errStr := dao.DeleteServerNode(in)
	if errCode != 0 {
		errStack := hcc_errors.NewHccErrorStack(hcc_errors.NewHccError(errCode, errStr))
		return &pb.ResDeleteServerNode{ServerNode: &pb.ServerNode{}, HccErrorStack: errconv.HccStackToGrpc(errStack)}, nil
	}

	return &pb.ResDeleteServerNode{ServerNode: deleteServerNode}, nil
}

func (s *violinServer) DeleteServerNodeByServerUUID(_ context.Context, in *pb.ReqDeleteServerNodeByServerUUID) (*pb.ResDeleteServerNodeByServerUUID, error) {
	logger.Logger.Println("Request received: DeleteServerNodeByServerUUID()")

	serverUUID, errCode, errStr := dao.DeleteServerNodeByServerUUID(in)
	if errCode != 0 {
		errStack := hcc_errors.NewHccErrorStack(hcc_errors.NewHccError(errCode, errStr))
		return &pb.ResDeleteServerNodeByServerUUID{ServerUUID: "", HccErrorStack: errconv.HccStackToGrpc(errStack)}, nil
	}

	return &pb.ResDeleteServerNodeByServerUUID{ServerUUID: serverUUID}, nil
}
