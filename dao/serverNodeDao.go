package dao

import (
	"errors"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"hcc/violin/model"
	"time"

	gouuid "github.com/nu7hatch/gouuid"
)

// ReadServerNode - cgs
func ReadServerNode(args map[string]interface{}) (interface{}, error) {
	var serverNode model.ServerNode
	var err error
	uuid := args["uuid"].(string)
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
	serverNode.CreatedAt = createdAt

	return serverNode, nil
}

// ReadServerNodeList - cgs
func ReadServerNodeList(args map[string]interface{}) (interface{}, error) {
	var err error
	var serverNodes []model.ServerNode
	var uuid string
	var nodeUUID string
	var createdAt time.Time
	serverUUID, serverUUIDOk := args["server_uuid"].(string)

	if !serverUUIDOk {
		return nil, err
	}

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
		serverNode := model.ServerNode{UUID: uuid, ServerUUID: serverUUID, NodeUUID: nodeUUID, CreatedAt: createdAt}
		logger.Logger.Println(serverNode)
		serverNodes = append(serverNodes, serverNode)
	}
	return serverNodes, nil
}

// ReadServerNodeAll - cgs
func ReadServerNodeAll() (interface{}, error) {
	var err error
	var serverNodes []model.ServerNode
	var uuid string
	var serverUUID string
	var nodeUUID string
	var createdAt time.Time

	sql := "select * from server_node order by created_at desc"
	stmt, err := mysql.Db.Query(sql)
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
			logger.Logger.Println(err)
			return nil, err
		}
		serverNode := model.ServerNode{UUID: uuid, ServerUUID: serverUUID, NodeUUID: nodeUUID, CreatedAt: createdAt}
		serverNodes = append(serverNodes, serverNode)
	}

	return serverNodes, nil
}

// ReadServerNodeNum - ish
func ReadServerNodeNum(args map[string]interface{}) (interface{}, error) {
	logger.Logger.Println("serverNodeDao: ReadServerNodeNum")
	var serverNodeNum model.ServerNodeNum
	var serverNodeNr int
	var err error

	serverUUID, serverUUIDOk := args["server_uuid"].(string)
	if !serverUUIDOk {
		return nil, errors.New("need a server_uuid argument")
	}

	sql := "select count(*) from server_node where server_uuid = '" + serverUUID + "'"
	err = mysql.Db.QueryRow(sql).Scan(&serverNodeNr)
	if err != nil {
		logger.Logger.Println(err)
		return nil, err
	}
	serverNodeNum.Number = serverNodeNr

	return serverNodeNum, nil
}


// CreateServerNode - cgs
func CreateServerNode(args map[string]interface{}) (interface{}, error) {
	out, err := gouuid.NewV4()
	if err != nil {
		logger.Logger.Println(err)
		return nil, err
	}
	uuid := out.String()

	serverNode := model.ServerNode{
		UUID:       uuid,
		ServerUUID: args["server_uuid"].(string),
		NodeUUID:   args["node_uuid"].(string),
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

	return serverNode, nil
}

// DeleteServerNode - cgs
func DeleteServerNode(args map[string]interface{}) (interface{}, error) {
	var err error

	requestedUUID, ok := args["uuid"].(string)
	if ok {
		sql := "delete from server_node where uuid = ?"
		stmt, err := mysql.Db.Prepare(sql)
		if err != nil {
			logger.Logger.Println(err.Error())
			return nil, err
		}
		defer func() {
			_ = stmt.Close()
		}()
		result, err2 := stmt.Exec(requestedUUID)
		if err2 != nil {
			logger.Logger.Println(err2)
			return nil, err
		}
		logger.Logger.Println(result.RowsAffected())

		return requestedUUID, nil
	}

	return requestedUUID, err
}
