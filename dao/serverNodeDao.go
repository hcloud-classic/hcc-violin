package dao

import (
	"errors"
	"github.com/golang/protobuf/ptypes"
	gouuid "github.com/nu7hatch/gouuid"
	pb "hcc/violin/action/grpc/rpcviolin"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"time"
)

// ReadServerNode : Get infos of a server node
func ReadServerNode(uuid string) (*pb.ServerNode, error) {
	var serverNode pb.ServerNode
	var err error
	var serverUUID string
	var nodeUUID string
	var createdAt time.Time

	sql := "select * from server_node where uuid = ?"
	err = mysql.Db.QueryRow(sql, uuid).Scan(
		&uuid,
		&serverUUID,
		&nodeUUID,
		&createdAt)
	if err != nil {
		logger.Logger.Println(err)
		return nil, err
	}

	serverNode.UUID = uuid
	serverNode.ServerUUID = serverUUID
	serverNode.NodeUUID = nodeUUID

	serverNode.CreatedAt, err = ptypes.TimestampProto(createdAt)
	if err != nil {
		logger.Logger.Println(err)
		return nil, err
	}

	return &serverNode, nil
}

// ReadServerNodeList : Get list of server nodes with provided server UUID
func ReadServerNodeList(in *pb.ReqGetServerNodeList) (*pb.ResGetServerNodeList, error) {
	serverUUID := in.GetServerUUID()
	serverUUIDOk := len(serverUUID) != 0
	if !serverUUIDOk {
		return nil, errors.New("need a serverUUID argument")
	}

	var serverNodeList pb.ResGetServerNodeList
	var serverNodes []pb.ServerNode
	var pserverNodes []*pb.ServerNode

	var uuid string
	var nodeUUID string
	var createdAt time.Time

	sql := "select * from server_node where server_uuid = ?"

	stmt, err := mysql.Db.Query(sql, serverUUID)
	if err != nil {
		logger.Logger.Println(err.Error())
		return nil, err
	}
	defer func() {
		_ = stmt.Close()
	}()

	for stmt.Next() {
		err := stmt.Scan(&uuid, &serverUUID, &nodeUUID, &createdAt)
		if err != nil {
			logger.Logger.Println(err.Error())
			return nil, err
		}

		_createdAt, err := ptypes.TimestampProto(createdAt)
		if err != nil {
			logger.Logger.Println(err)
			return nil, err
		}

		serverNodes = append(serverNodes, pb.ServerNode{
			UUID:       uuid,
			ServerUUID: serverUUID,
			NodeUUID:   nodeUUID,
			CreatedAt:  _createdAt})
	}

	for i := range serverNodes {
		pserverNodes = append(pserverNodes, &serverNodes[i])
	}

	serverNodeList.ServerNodeList = pserverNodes

	return &serverNodeList, nil
}

// ReadServerNodeNum : Get the number of server nodes
func ReadServerNodeNum(in *pb.ReqGetServerNodeNum) (*pb.ResGetServerNodeNum, error) {
	serverUUID := in.GetServerUUID()
	serverUUIDOk := len(serverUUID) != 0
	if !serverUUIDOk {
		return nil, errors.New("need a serverUUID argument")
	}

	var serverNodeNum pb.ResGetServerNodeNum
	var serverNodeNr int
	var err error

	sql := "select count(*) from server_node where server_uuid = '" + serverUUID + "'"
	err = mysql.Db.QueryRow(sql).Scan(&serverNodeNr)
	if err != nil {
		logger.Logger.Println(err)
		return nil, err
	}
	serverNodeNum.Num = int64(serverNodeNr)

	return &serverNodeNum, nil
}

func checkCreateServerNodeArgs(reqServerNode *pb.ServerNode) bool {
	serverUUIDOk := len(reqServerNode.ServerUUID) != 0
	nodeUUIDOk := len(reqServerNode.NodeUUID) != 0

	return !(serverUUIDOk && nodeUUIDOk)
}

// CreateServerNode : Create server nodes. Insert each node UUIDs with server UUID.
func CreateServerNode(in *pb.ReqCreateServerNode) (*pb.ServerNode, error) {
	reqServerNode := in.GetServerNode()
	if reqServerNode == nil {
		return nil, errors.New("serverNode is nil")
	}

	out, err := gouuid.NewV4()
	if err != nil {
		logger.Logger.Println(err)
		return nil, err
	}
	uuid := out.String()

	if checkCreateServerNodeArgs(reqServerNode) {
		return nil, errors.New("some of arguments are missing")
	}

	serverNodeList, err := ReadServerNodeList(&pb.ReqGetServerNodeList{ServerUUID: reqServerNode.ServerUUID})
	if err != nil {
		return nil, err
	}
	pserverNodes := serverNodeList.ServerNodeList

	for i := range pserverNodes {
		if pserverNodes[i].NodeUUID == reqServerNode.NodeUUID {
			return nil, errors.New("requested ServerNode is already present in the database (" +
				"UUID: " + pserverNodes[i].UUID + ", " +
				"ServerUUID: " + pserverNodes[i].ServerUUID + ", " +
				"NodeUUID: " + pserverNodes[i].NodeUUID + ")")
		}
	}

	serverNode := pb.ServerNode{
		UUID:       uuid,
		ServerUUID: reqServerNode.ServerUUID,
		NodeUUID:   reqServerNode.NodeUUID,
	}

	sql := "insert into server_node(uuid, server_uuid, node_uuid, created_at) values (?, ?, ?, now())"
	stmt, err := mysql.Db.Prepare(sql)
	if err != nil {
		logger.Logger.Println(err)
		return nil, err
	}
	defer func() {
		_ = stmt.Close()
	}()

	result, err := stmt.Exec(serverNode.UUID, serverNode.ServerUUID, serverNode.NodeUUID)
	if err != nil {
		logger.Logger.Println(err)
		return nil, err
	}
	logger.Logger.Println(result.LastInsertId())

	return &serverNode, nil
}

// DeleteServerNode : Delete server nodes. Delete server nodes matched with server UUID.
func DeleteServerNode(in *pb.ReqDeleteServerNode) (string, error) {
	var err error

	requestedUUID := in.GetServerUUID()
	requestedUUIDOk := len(requestedUUID) != 0
	if !requestedUUIDOk {
		return "", errors.New("need a serverUUID argument")
	}

	sql := "delete from server_node where server_uuid = ?"
	stmt, err := mysql.Db.Prepare(sql)
	if err != nil {
		logger.Logger.Println(err.Error())
		return "", err
	}
	defer func() {
		_ = stmt.Close()
	}()
	result, err2 := stmt.Exec(requestedUUID)
	if err2 != nil {
		logger.Logger.Println(err2)
		return "", err
	}
	logger.Logger.Println(result.RowsAffected())

	return requestedUUID, nil
}
