package dao

import (
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"hcc/violin/lib/uuidgen"
	"hcc/violin/model"
	"time"
)

// Read

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
		logger.Log.Println(err)
		return nil, err
	}

	serverNode.UUID = uuid
	serverNode.ServerUUID = serverUUID
	serverNode.NodeUUID = nodeUUID
	serverNode.CreatedAt = createdAt

	return serverNode, nil
}

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
		logger.Log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	for stmt.Next() {
		err := stmt.Scan(&uuid, &serverUUID, &nodeUUID, &createdAt)
		if err != nil {
			logger.Log.Println(err.Error())
			return nil, err
		}
		serverNode := model.ServerNode{UUID: uuid, ServerUUID: serverUUID, NodeUUID: nodeUUID, CreatedAt: createdAt}
		logger.Log.Println(serverNode)
		serverNodes = append(serverNodes, serverNode)
	}
	return serverNodes, nil
}

func ReadServerNodeAll(args map[string]interface{}) (interface{}, error) {
	var err error
	var serverNodes []model.ServerNode
	var uuid string
	var serverUUID string
	var nodeUUID string
	var createdAt time.Time

	sql := "select * from server_node order by created_at desc"
	stmt, err := mysql.Db.Query(sql)
	if err != nil {
		logger.Log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	for stmt.Next() {
		err := stmt.Scan(&uuid, &serverUUID, &nodeUUID, &createdAt)
		if err != nil {
			logger.Log.Println(err)
			return nil, err
		}
		serverNode := model.ServerNode{UUID: uuid, ServerUUID: serverUUID, NodeUUID: nodeUUID, CreatedAt: createdAt}
		serverNodes = append(serverNodes, serverNode)
	}

	return serverNodes, nil
}

// Create

func CreateServerNode(args map[string]interface{}) (interface{}, error) {
	uuid, err := uuidgen.Uuidgen()
	if err != nil {
		logger.Log.Println("Failed to generate uuid!")
		return nil, err
	}

	serverNode := model.ServerNode{
		UUID:       uuid,
		ServerUUID: args["server_uuid"].(string),
		NodeUUID:   args["node_uuid"].(string),
	}

	sql := "insert into server_node(uuid, server_uuid, node_uuid, created_at) values (?, ?, ?, now())"
	stmt, err := mysql.Db.Prepare(sql)
	if err != nil {
		logger.Log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()
	result, err := stmt.Exec(serverNode.UUID, serverNode.ServerUUID, serverNode.NodeUUID)
	if err != nil {
		logger.Log.Println(err)
		return nil, err
	}
	logger.Log.Println(result.LastInsertId())

	return serverNode, nil
}

// Update

// Delete

func DeleteServerNode(args map[string]interface{}) (interface{}, error) {
	var err error

	requestedUUID, ok := args["uuid"].(string)
	if ok {
		sql := "delete from server_node where uuid = ?"
		stmt, err := mysql.Db.Prepare(sql)
		if err != nil {
			logger.Log.Println(err.Error())
			return nil, err
		}
		defer stmt.Close()
		result, err2 := stmt.Exec(requestedUUID)
		if err2 != nil {
			logger.Log.Println(err2)
			return nil, err
		}
		logger.Log.Println(result.RowsAffected())

		return requestedUUID, nil
	}

	return requestedUUID, err
}
