package dao

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"innogrid.com/hcloud-classic/hcc_errors"
	"innogrid.com/hcloud-classic/pb"
	"strings"
	"time"
)

// ReadServerNode : Get infos of a server node
func ReadServerNode(uuid string) (*pb.ServerNode, uint64, string) {
	var serverNode pb.ServerNode
	var err error
	var serverUUID string
	var nodeUUID string
	var createdAt time.Time

	sql := "select * from server_node where uuid = ?"
	row := mysql.Db.QueryRow(sql, uuid)
	err = mysql.QueryRowScan(row,
		&uuid,
		&serverUUID,
		&nodeUUID,
		&createdAt)
	if err != nil {
		errStr := "ReadServerNode(): " + err.Error()
		logger.Logger.Println(errStr)
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, hcc_errors.ViolinSQLNoResult, errStr
		}
		return nil, hcc_errors.ViolinSQLOperationFail, errStr
	}

	serverNode.UUID = uuid
	serverNode.ServerUUID = serverUUID
	serverNode.NodeUUID = nodeUUID
	serverNode.CreatedAt = timestamppb.New(createdAt)

	return &serverNode, 0, ""
}

// ReadServerNodeNum : Get the number of server nodes
func ReadServerNodeNum(in *pb.ReqGetServerNodeNum) (*pb.ResGetServerNodeNum, uint64, string) {
	serverUUID := in.GetServerUUID()
	serverUUIDOk := len(serverUUID) != 0
	if !serverUUIDOk {
		return nil, hcc_errors.ViolinGrpcArgumentError, "ReadServerNodeNum(): need a serverUUID argument"
	}

	var serverNodeNum pb.ResGetServerNodeNum
	var serverNodeNr int
	var err error

	sql := "select count(*) from server_node where server_uuid = '" + serverUUID + "'"
	row := mysql.Db.QueryRow(sql)
	err = mysql.QueryRowScan(row, &serverNodeNr)
	if err != nil {
		errStr := "ReadServerNodeNum(): " + err.Error()
		logger.Logger.Println(errStr)
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, hcc_errors.ViolinSQLNoResult, errStr
		}
		return nil, hcc_errors.ViolinSQLOperationFail, errStr
	}
	serverNodeNum.Num = int64(serverNodeNr)

	return &serverNodeNum, 0, ""
}

// DeleteServerNode : Delete a server node matched with provided UUID.
func DeleteServerNode(in *pb.ReqDeleteServerNode) (*pb.ServerNode, uint64, string) {
	var err error

	requestedUUID := in.GetUUID()
	requestedUUIDOk := len(requestedUUID) != 0
	if !requestedUUIDOk {
		return nil, hcc_errors.ViolinGrpcArgumentError, "DeleteServerNode(): need a UUID argument"
	}

	serverNode, errCode, errText := ReadServerNode(requestedUUID)
	if errCode != 0 {
		return nil, hcc_errors.ViolinGrpcRequestError, "DeleteServerNode(): " + errText
	}

	sql := "delete from server_node where uuid = ?"
	stmt, err := mysql.Prepare(sql)
	if err != nil {
		errStr := "DeleteServerNode(): " + err.Error()
		logger.Logger.Println(errStr)
		return nil, hcc_errors.ViolinSQLOperationFail, errStr
	}
	defer func() {
		_ = stmt.Close()
	}()
	result, err2 := stmt.Exec(requestedUUID)
	if err2 != nil {
		errStr := "DeleteServerNode(): " + err2.Error()
		logger.Logger.Println(errStr)
		return nil, hcc_errors.ViolinSQLOperationFail, errStr
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
		return "", hcc_errors.ViolinGrpcArgumentError, "DeleteServerNodeByServerUUID(): need a serverUUID argument"
	}

	sql := "delete from server_node where server_uuid = ?"
	stmt, err := mysql.Prepare(sql)
	if err != nil {
		errStr := "DeleteServerNodeByServerUUID(): " + err.Error()
		logger.Logger.Println(errStr)
		return "", hcc_errors.ViolinSQLOperationFail, errStr
	}
	defer func() {
		_ = stmt.Close()
	}()
	result, err2 := stmt.Exec(requestedServerUUID)
	if err2 != nil {
		errStr := "DeleteServerNodeByServerUUID(): " + err2.Error()
		logger.Logger.Println(errStr)
		return "", hcc_errors.ViolinSQLOperationFail, errStr
	}
	logger.Logger.Println(result.RowsAffected())

	return requestedServerUUID, 0, ""
}
