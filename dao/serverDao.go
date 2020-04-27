package dao

import (
	dbsql "database/sql"
	"errors"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"hcc/violin/model"
	"strconv"
	"time"
)

// ReadServer - cgs
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

// ReadServerList - cgs
func ReadServerList(args map[string]interface{}) (interface{}, error) {
	var servers []model.Server
	var rxUUID string
	var createdAt time.Time

	subnetUUID, subnetUUIDOk := args["subnet_uuid"].(string)
	os, osOk := args["os"].(string)
	serverName, serverNameOk := args["server_name"].(string)
	serverDesc, serverDescOk := args["server_desc"].(string)
	cpu, cpuOk := args["cpu"].(int)
	memory, memoryOk := args["memory"].(int)
	diskSize, diskSizeOk := args["disk_size"].(int)
	status, statusOk := args["status"].(string)
	userUUID, userUUIDOk := args["user_uuid"].(string)

	row, rowOk := args["row"].(int)
	page, pageOk := args["page"].(int)
	if !rowOk || !pageOk {
		return nil, errors.New("need row and page arguments")
	}

	sql := "select * from server where 1=1"
	if subnetUUIDOk {
		sql += " and subnet_uuid = '" + subnetUUID + "'"
	}
	if osOk {
		sql += " and os = '" + os + "'"
	}
	if serverNameOk {
		sql += " and server_name = '" + serverName + "'"
	}
	if serverDescOk {
		sql += " and server_desc = '" + serverDesc + "'"
	}
	if cpuOk {
		sql += " and cpu = " + strconv.Itoa(cpu)
	}
	if memoryOk {
		sql += " and memory = " + strconv.Itoa(memory)
	}
	if diskSizeOk {
		sql += " and disk_size = " + strconv.Itoa(diskSize)
	}
	if statusOk {
		sql += " and status = '" + status + "'"
	}
	if userUUIDOk {
		sql += " and user_uuid = '" + userUUID + "'"
	}

	sql += " order by created_at desc limit ? offset ?"
	logger.Logger.Println("list_server sql  : ", sql)

	stmt, err := mysql.Db.Query(sql, row, row*(page-1))
	if err != nil {
		logger.Logger.Println(err.Error())
		return nil, err
	}
	defer func() {
		_ = stmt.Close()
	}()

	for stmt.Next() {
		err := stmt.Scan(&rxUUID, &subnetUUID, &os, &serverName, &serverDesc, &cpu, &memory, &diskSize, &status, &userUUID, &createdAt)
		if err != nil {
			logger.Logger.Println(err.Error())
			return nil, err
		}
		server := model.Server{UUID: rxUUID, SubnetUUID: subnetUUID, OS: os, ServerName: serverName, ServerDesc: serverDesc, CPU: cpu, Memory: memory, DiskSize: diskSize, Status: status, UserUUID: userUUID, CreatedAt: createdAt}
		logger.Logger.Println(server)
		servers = append(servers, server)
	}
	return servers, nil
}

// ReadServerAll - cgs
func ReadServerAll(args map[string]interface{}) (interface{}, error) {
	var servers []model.Server
	var uuid string
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

	row, rowOk := args["row"].(int)
	page, pageOk := args["page"].(int)
	var sql string
	var stmt *dbsql.Rows
	var err error

	if !rowOk && !pageOk {
		sql = "select * from server order by created_at desc"
		stmt, err = mysql.Db.Query(sql)
	} else if rowOk && pageOk {
		sql = "select * from server order by created_at desc limit ? offset ?"
		stmt, err = mysql.Db.Query(sql, row, row*(page-1))
	} else {
		return nil, errors.New("please insert row and page arguments or leave arguments as empty state")
	}

	logger.Logger.Println("list_server sql  : ", sql)

	if err != nil {
		logger.Logger.Println(err.Error())
		return nil, err
	}
	defer func() {
		_ = stmt.Close()
	}()

	for stmt.Next() {
		err := stmt.Scan(&uuid, &subnetUUID, &os, &serverName, &serverDesc, &cpu, &memory, &diskSize, &status, &userUUID, &createdAt)
		if err != nil {
			logger.Logger.Println(err)
			return nil, err
		}
		server := model.Server{UUID: uuid, SubnetUUID: subnetUUID, OS: os, ServerName: serverName, ServerDesc: serverDesc, CPU: cpu, Memory: memory, DiskSize: diskSize, Status: status, UserUUID: userUUID, CreatedAt: createdAt}
		servers = append(servers, server)
	}

	return servers, nil
}

// ReadServerNum - cgs
func ReadServerNum() (model.ServerNum, error) {
	logger.Logger.Println("serverDao: ReadServerNum")
	var serverNum model.ServerNum
	var serverNr int
	var err error

	sql := "select count(*) from server where status != 'Deleted'"
	err = mysql.Db.QueryRow(sql).Scan(&serverNr)
	if err != nil {
		logger.Logger.Println(err)
		return serverNum, err
	}
	serverNum.Number = serverNr

	return serverNum, nil
}

// CreateServer - cgs
func CreateServer(serverUUID string, args map[string]interface{}) (interface{}, error) {
	server := model.Server{
		UUID:       serverUUID,
		SubnetUUID: args["subnet_uuid"].(string),
		OS:         args["os"].(string),
		ServerName: args["server_name"].(string),
		ServerDesc: args["server_desc"].(string),
		CPU:        args["cpu"].(int),
		Memory:     args["memory"].(int),
		DiskSize:   args["disk_size"].(int),
		Status:     "Creating",
		UserUUID:   args["user_uuid"].(string),
	}

	sql := "insert into server(uuid, subnet_uuid, os, server_name, server_desc, cpu, memory, disk_size, status, user_uuid, created_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now())"
	stmt, err := mysql.Db.Prepare(sql)
	if err != nil {
		logger.Logger.Println(err.Error())
		return nil, err
	}
	defer func() {
		_ = stmt.Close()
	}()
	result, err := stmt.Exec(server.UUID, server.SubnetUUID, server.OS, server.ServerName, server.ServerDesc, server.CPU, server.Memory, server.DiskSize, server.Status, server.UserUUID)
	if err != nil {
		logger.Logger.Println(err)
		return nil, err
	}
	logger.Logger.Println(result.LastInsertId())

	return server, nil
}

func checkUpdateServerArgs(args map[string]interface{}) bool {
	_, subnetUUIDOk := args["subnet_uuid"].(string)
	_, osOk := args["os"].(string)
	_, serverNameOk := args["server_name"].(string)
	_, serverDescOk := args["server_desc"].(string)
	_, cpuOk := args["cpu"].(int)
	_, memoryOk := args["memory"].(int)
	_, diskSizeOk := args["disk_size"].(int)
	_, statusOk := args["status"].(string)
	_, userUUIDOk := args["user_uuid"].(string)

	return !subnetUUIDOk && !osOk && !serverNameOk && !serverDescOk && !cpuOk && !memoryOk && !diskSizeOk && !statusOk && !userUUIDOk
}

// UpdateServer - cgs
func UpdateServer(args map[string]interface{}) (interface{}, error) {
	requestedUUID, requestedUUIDOk := args["uuid"].(string)
	subnetUUID, subnetUUIDOk := args["subnet_uuid"].(string)
	os, osOk := args["os"].(string)
	serverName, serverNameOk := args["server_name"].(string)
	serverDesc, serverDescOk := args["server_desc"].(string)
	cpu, cpuOk := args["cpu"].(int)
	memory, memoryOk := args["memory"].(int)
	diskSize, diskSizeOk := args["disk_size"].(int)
	status, statusOk := args["status"].(string)
	userUUID, userUUIDOk := args["user_uuid"].(string)

	server := new(model.Server)
	server.UUID = requestedUUID
	server.SubnetUUID = subnetUUID
	server.OS = os
	server.ServerName = serverName
	server.ServerDesc = serverDesc
	server.CPU = cpu
	server.Memory = memory
	server.DiskSize = diskSize
	server.Status = status
	server.UserUUID = userUUID

	if requestedUUIDOk {
		if checkUpdateServerArgs(args) {
			return nil, errors.New("need some arguments")
		}

		sql := "update server set"
		var updateSet = ""
		if subnetUUIDOk {
			updateSet += " subnet_uuid = '" + server.SubnetUUID + "', "
		}
		if osOk {
			updateSet += " os = '" + server.OS + "', "
		}
		if serverNameOk {
			updateSet += " server_name = '" + server.ServerName + "', "
		}
		if serverDescOk {
			updateSet += " server_desc = '" + server.ServerDesc + "', "
		}
		if cpuOk {
			updateSet += " cpu = " + strconv.Itoa(server.CPU) + ", "
		}
		if memoryOk {
			updateSet += " memory = " + strconv.Itoa(server.Memory) + ", "
		}
		if diskSizeOk {
			updateSet += " disk_size = " + strconv.Itoa(server.DiskSize) + ", "
		}
		if statusOk {
			updateSet += " status = '" + server.Status + "', "
		}
		if userUUIDOk {
			updateSet += " user_uuid = " + server.UserUUID + "', "
		}

		sql += updateSet[0:len(updateSet)-2] + " where uuid = ?"

		logger.Logger.Println("update_server sql : ", sql)

		stmt, err := mysql.Db.Prepare(sql)
		if err != nil {
			logger.Logger.Println(err.Error())
			return nil, err
		}
		defer func() {
			_ = stmt.Close()
		}()

		result, err2 := stmt.Exec(server.UUID)
		if err2 != nil {
			logger.Logger.Println(err2)
			return nil, err
		}
		logger.Logger.Println(result.LastInsertId())
		return server, nil
	}

	return nil, errors.New("uuid argument is missing")
}

// UpdateServerStatus : Update status of the server
func UpdateServerStatus(server_uuid string, status string) error {
	sql := "update server set status = '" + status + "' where uuid = ?"

	logger.Logger.Println("UpdateServerStatus sql : ", sql)

	stmt, err := mysql.Db.Prepare(sql)
	if err != nil {
		logger.Logger.Println(err.Error())
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()

	result, err2 := stmt.Exec(server_uuid)
	if err2 != nil {
		logger.Logger.Println(err2)
		return err
	}
	logger.Logger.Println(result.LastInsertId())
	return nil
}

// DeleteServer - cgs
func DeleteServer(args map[string]interface{}) (interface{}, error) {
	var err error

	var server model.Server
	requestedUUID, ok := args["uuid"].(string)
	server.UUID = requestedUUID
	if ok {
		sql := "delete from server where uuid = ?"
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
			return nil, err2
		}
		logger.Logger.Println(result.RowsAffected())

		return server, nil
	}

	return nil, err
}
