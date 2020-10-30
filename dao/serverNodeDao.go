package dao

import (
	pb "hcc/violin/action/grpc/pb/rpcviolin"
	hccerr "hcc/violin/lib/errors"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	gouuid "github.com/nu7hatch/gouuid"
)

// ReadServerNode : Get infos of a server node
func ReadServerNode(uuid string) (*pb.ServerNode, uint64, string) {
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
		errStr := "ReadServerNode(): " + err.Error()
		logger.Logger.Println(errStr)
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, hccerr.ViolinSQLNoResult, errStr
		}
		return nil, hccerr.ViolinSQLOperationFail, errStr
	}

	serverNode.UUID = uuid
	serverNode.ServerUUID = serverUUID
	serverNode.NodeUUID = nodeUUID

	serverNode.CreatedAt, err = ptypes.TimestampProto(createdAt)
	if err != nil {
		errStr := "ReadServerNode(): " + err.Error()
		logger.Logger.Println(errStr)
		return nil, hccerr.ViolinInternalTimeStampConversionError, errStr
	}

	return &serverNode, 0, ""
}

// ReadServerNodeList : Get list of server nodes with provided server UUID
func ReadServerNodeList(in *pb.ReqGetServerNodeList) (*pb.ResGetServerNodeList, uint64, string) {
	serverUUID := in.GetServerUUID()
	serverUUIDOk := len(serverUUID) != 0
	if !serverUUIDOk {
		return nil, hccerr.ViolinGrpcArgumentError, "ReadServerNodeList(): need a serverUUID argument"
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
		errStr := "ReadServerNodeList(): " + err.Error()
		logger.Logger.Println(errStr)
		return nil, hccerr.ViolinSQLOperationFail, errStr
	}
	defer func() {
		_ = stmt.Close()
	}()

	for stmt.Next() {
		err := stmt.Scan(&uuid, &serverUUID, &nodeUUID, &createdAt)
		if err != nil {
			errStr := "ReadServerNodeList(): " + err.Error()
			logger.Logger.Println(errStr)
			if strings.Contains(err.Error(), "no rows in result set") {
				return nil, hccerr.ViolinSQLNoResult, errStr
			}
			return nil, hccerr.ViolinSQLOperationFail, errStr
		}

		_createdAt, err := ptypes.TimestampProto(createdAt)
		if err != nil {
			errStr := "ReadServerNodeList(): " + err.Error()
			logger.Logger.Println(errStr)
			return nil, hccerr.ViolinInternalTimeStampConversionError, errStr
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

	serverNodeList.ServerNode = pserverNodes

	return &serverNodeList, 0, ""
}

// ReadServerNodeNum : Get the number of server nodes
func ReadServerNodeNum(in *pb.ReqGetServerNodeNum) (*pb.ResGetServerNodeNum, uint64, string) {
	serverUUID := in.GetServerUUID()
	serverUUIDOk := len(serverUUID) != 0
	if !serverUUIDOk {
		return nil, hccerr.ViolinGrpcArgumentError, "ReadServerNodeNum(): need a serverUUID argument"
	}

	var serverNodeNum pb.ResGetServerNodeNum
	var serverNodeNr int
	var err error

	sql := "select count(*) from server_node where server_uuid = '" + serverUUID + "'"
	err = mysql.Db.QueryRow(sql).Scan(&serverNodeNr)
	if err != nil {
		errStr := "ReadServerNodeNum(): " + err.Error()
		logger.Logger.Println(errStr)
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, hccerr.ViolinSQLNoResult, errStr
		}
		return nil, hccerr.ViolinSQLOperationFail, errStr
	}
	serverNodeNum.Num = int64(serverNodeNr)

	return &serverNodeNum, 0, ""
}

func checkCreateServerNodeArgs(reqServerNode *pb.ServerNode) bool {
	serverUUIDOk := len(reqServerNode.ServerUUID) != 0
	nodeUUIDOk := len(reqServerNode.NodeUUID) != 0
	return !(serverUUIDOk && nodeUUIDOk)
}

// CreateServerNode : Create server nodes. Insert each node UUIDs with server UUID.
func CreateServerNode(in *pb.ReqCreateServerNode) (*pb.ServerNode, uint64, string) {
	reqServerNode := in.GetServerNode()
	if reqServerNode == nil {
		return nil, hccerr.ViolinGrpcArgumentError, "CreateServerNode(): serverNode is nil"
	}

	out, err := gouuid.NewV4()
	if err != nil {
		logger.Logger.Println(err)
		return nil, hccerr.ViolinInternalUUIDGenerationError, "CreateServerNode(): " + err.Error()
	}
	uuid := out.String()

	if checkCreateServerNodeArgs(reqServerNode) {
		return nil, hccerr.ViolinGrpcArgumentError, "CreateServerNode(): some of arguments are missing\n"
	}

	serverNodeList, errCode, errStr := ReadServerNodeList(&pb.ReqGetServerNodeList{ServerUUID: reqServerNode.ServerUUID})
	if errCode != 0 {
		return nil, errCode, "CreateServerNode(): " + errStr
	}
	pserverNodes := serverNodeList.ServerNode

	for i := range pserverNodes {
		if pserverNodes[i].NodeUUID == reqServerNode.NodeUUID {
			return nil, hccerr.ViolinInternalServerNodePresentError,
				"CreateServerNode(): requested ServerNode is already present in the database (" +
					"UUID: " + pserverNodes[i].UUID + ", " +
					"ServerUUID: " + pserverNodes[i].ServerUUID + ", " +
					"NodeUUID: " + pserverNodes[i].NodeUUID + ")"
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
		errStr := "CreateServerNode(): " + err.Error()
		logger.Logger.Println(errStr)
		return nil, hccerr.ViolinSQLOperationFail, errStr
	}
	defer func() {
		_ = stmt.Close()
	}()

	_, err = stmt.Exec(serverNode.UUID, serverNode.ServerUUID, serverNode.NodeUUID)
	if err != nil {
		errStr := "CreateServerNode(): " + err.Error()
		logger.Logger.Println(errStr)
		return nil, hccerr.ViolinSQLOperationFail, errStr
	}

	return &serverNode, 0, ""
}

// DeleteServerNode : Delete a server node matched with provided UUID.
func DeleteServerNode(in *pb.ReqDeleteServerNode) (*pb.ServerNode, uint64, string) {
	var err error

	requestedUUID := in.GetUUID()
	requestedUUIDOk := len(requestedUUID) != 0
	if !requestedUUIDOk {
		return nil, hccerr.ViolinGrpcArgumentError, "DeleteServerNode(): need a UUID argument"
	}

	serverNode, errCode, errText := ReadServerNode(requestedUUID)
	if errCode != 0 {
		return nil, hccerr.ViolinGrpcRequestError, "DeleteServerNode(): " + errText
	}

	sql := "delete from server_node where uuid = ?"
	stmt, err := mysql.Db.Prepare(sql)
	if err != nil {
		errStr := "DeleteServerNode(): " + err.Error()
		logger.Logger.Println(errStr)
		return nil, hccerr.ViolinSQLOperationFail, errStr
	}
	defer func() {
		_ = stmt.Close()
	}()
	result, err2 := stmt.Exec(requestedUUID)
	if err2 != nil {
		errStr := "DeleteServerNode(): " + err2.Error()
		logger.Logger.Println(errStr)
		return nil, hccerr.ViolinSQLOperationFail, errStr
	}
	logger.Logger.Println(result.RowsAffected())

	return serverNode, 0, ""
}

// DeleteServerNodeByServerUUID : Delete server nodes. Delete server nodes matched with server UUID.
func DeleteServerNodeByServerUUID(in *pb.ReqDeleteServerNodeByServerUUID) (string, uint64, string) {
	var err error

	requestedServerUUID := in.GetServerUUID()
	requestedServerUUIDOk := len(requestedServerUUID) != 0
	if !requestedServerUUIDOk {
		return "", hccerr.ViolinGrpcArgumentError, "DeleteServerNodeByServerUUID(): need a serverUUID argument"
	}

	sql := "delete from server_node where server_uuid = ?"
	stmt, err := mysql.Db.Prepare(sql)
	if err != nil {
		errStr := "DeleteServerNodeByServerUUID(): " + err.Error()
		logger.Logger.Println(errStr)
		return "", hccerr.ViolinSQLOperationFail, errStr
	}
	defer func() {
		_ = stmt.Close()
	}()
	result, err2 := stmt.Exec(requestedServerUUID)
	if err2 != nil {
		errStr := "DeleteServerNodeByServerUUID(): " + err2.Error()
		logger.Logger.Println(errStr)
		return "", hccerr.ViolinSQLOperationFail, errStr
	}
	logger.Logger.Println(result.RowsAffected())

	return requestedServerUUID, 0, ""
}
