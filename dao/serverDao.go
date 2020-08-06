package dao

import (
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"hcc/violin/model"
	"time"
)

func ReadServer(args map[string]interface{}) (interface{}, error) {
	var server model.Server
	var err error
	uuid := args["uuid"].(string)
	var subnetUUID string
	var os string
	var serverName string
	var serverDesc string
	var cpu int
	var memory int
	var diskSize int
	var status string
	var userUUID string
	var createdAt time.Time

	sql := "select * from server where uuid = ?"
	err = mysql.Db.QueryRow(sql, uuid).Scan(
		&uuid,
		&subnetUUID,
		&os,
		&serverName,
		&serverDesc,
		&cpu,
		&memory,
		&diskSize,
		&status,
		&userUUID,
		&createdAt)
	if err != nil {
		logger.Logger.Println(err)
		return nil, err
	}

	server.UUID = uuid
	server.SubnetUUID = subnetUUID
	server.OS = os
	server.ServerName = serverName
	server.ServerDesc = serverDesc
	server.CPU = cpu
	server.Memory = memory
	server.DiskSize = diskSize
	server.Status = status
	server.UserUUID = userUUID
	server.CreatedAt = createdAt

	return server, nil
}
