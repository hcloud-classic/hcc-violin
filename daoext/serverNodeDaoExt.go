package daoext

import (
	"errors"
	gouuid "github.com/nu7hatch/gouuid"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"innogrid.com/hcloud-classic/hcc_errors"
	"innogrid.com/hcloud-classic/pb"
)

// CreateServerNode : Create server nodes. Insert each node UUIDs with server UUID.
func CreateServerNode(in *pb.ReqCreateServerNode) (*pb.ServerNode, uint64, string) {
	reqServerNode := in.GetServerNode()
	if reqServerNode == nil {
		return nil, hcc_errors.ViolinGrpcArgumentError, "CreateServerNode(): serverNode is nil"
	}

	out, err := gouuid.NewV4()
	if err != nil {
		logger.Logger.Println(err)
		return nil, hcc_errors.ViolinInternalUUIDGenerationError, "CreateServerNode(): " + err.Error()
	}
	uuid := out.String()

	if CheckCreateServerNodeArgs(reqServerNode) {
		return nil, hcc_errors.ViolinGrpcArgumentError, "CreateServerNode(): some of arguments are missing\n"
	}

	serverNodeList, errCode, errStr := ReadServerNodeList(&pb.ReqGetServerNodeList{ServerUUID: reqServerNode.ServerUUID})
	if errCode != 0 {
		return nil, errCode, "CreateServerNode(): " + errStr
	}
	pserverNodes := serverNodeList.ServerNode

	for i := range pserverNodes {
		if pserverNodes[i].NodeUUID == reqServerNode.NodeUUID {
			return nil, hcc_errors.ViolinInternalServerNodePresentError,
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
	stmt, err := mysql.Prepare(sql)
	if err != nil {
		errStr := "CreateServerNode(): " + err.Error()
		logger.Logger.Println(errStr)
		return nil, hcc_errors.ViolinSQLOperationFail, errStr
	}
	defer func() {
		_ = stmt.Close()
	}()

	_, err = stmt.Exec(serverNode.UUID, serverNode.ServerUUID, serverNode.NodeUUID)
	if err != nil {
		errStr := "CreateServerNode(): " + err.Error()
		logger.Logger.Println(errStr)
		return nil, hcc_errors.ViolinSQLOperationFail, errStr
	}

	return &serverNode, 0, ""
}

// DeleteServerNodeByNodeUUID : Delete a server node matched with the node UUID.
func DeleteServerNodeByNodeUUID(nodeUUID string) error {
	if len(nodeUUID) == 0 {
		return errors.New("DeleteServerNodeByNodeUUID(): Please provide nodeUUID")
	}

	sql := "delete from server_node where node_uuid = ?"
	stmt, err := mysql.Prepare(sql)
	if err != nil {
		errStr := "DeleteServerNodeByNodeUUID(): " + err.Error()
		logger.Logger.Println(errStr)
		return errors.New(errStr)
	}
	defer func() {
		_ = stmt.Close()
	}()
	result, err2 := stmt.Exec(nodeUUID)
	if err2 != nil {
		errStr := "DeleteServerNodeByNodeUUID(): " + err2.Error()
		logger.Logger.Println(errStr)
		return errors.New(errStr)
	}
	logger.Logger.Println(result.RowsAffected())

	return nil
}
