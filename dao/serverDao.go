package dao

import (
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"hcc/violin/lib/uuidgen"
	"hcc/violin/model"
	"strconv"
	"time"
)

// ReadServer :
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

// ReadServerList :
func ReadServerList(args map[string]interface{}) (interface{}, error) {
	var err error
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

	if !userUUIDOk {
		return nil, err
	}
	row, rowOk := args["row"].(int)
	page, pageOk := args["page"].(int)
	if !rowOk || !pageOk {
		return nil, err
	}

	sql := "select * from server where"
	if subnetUUIDOk {
		sql += " subnet_uuid = '" + subnetUUID + "'"
		if osOk || serverNameOk || serverDescOk || cpuOk || memoryOk || diskSizeOk || statusOk || userUUIDOk {
			sql += " and"
		}
	}
	if osOk {
		sql += " os = '" + os + "'"
		if serverNameOk || serverDescOk || cpuOk || memoryOk || diskSizeOk || statusOk || userUUIDOk {
			sql += " and"
		}
	}
	if serverNameOk {
		sql += " server_name = '" + serverName + "'"
		if serverDescOk || cpuOk || memoryOk || diskSizeOk || statusOk || userUUIDOk {
			sql += " and"
		}
	}
	if serverDescOk {
		sql += " server_desc = '" + serverDesc + "'"
		if cpuOk || memoryOk || diskSizeOk || statusOk || userUUIDOk {
			sql += " and"
		}
	}
	if cpuOk {
		sql += " cpu = " + strconv.Itoa(cpu)
		if memoryOk || diskSizeOk || statusOk || userUUIDOk {
			sql += " and"
		}
	}
	if memoryOk {
		sql += " memory = " + strconv.Itoa(memory)
		if diskSizeOk || statusOk || userUUIDOk {
			sql += " and"
		}
	}
	if diskSizeOk {
		sql += " disk_size = " + strconv.Itoa(diskSize)
		if statusOk || userUUIDOk {
			sql += " and"
		}
	}
	if statusOk {
		sql += " status = '" + status + "' and"
	}

	sql += " user_uuid = ? order by created_at desc limit ? offset ?"
	logger.Logger.Println("list_server sql  : ", sql)

	stmt, err := mysql.Db.Query(sql, userUUID, row, row*(page-1))
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

// ReadServerAll :
func ReadServerAll(args map[string]interface{}) (interface{}, error) {
	var err error
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
	if !rowOk || !pageOk {
		return nil, err
	}

	sql := "select * from server order by created_at desc limit ? offset ?"
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

// ReadServerNum :
func ReadServerNum() (model.ServerNum, error) {
	logger.Logger.Println("serverDao: ReadServerNum")
	var serverNum model.ServerNum
	var serverNr int
	var err error

	sql := "select count(*) from server"
	err = mysql.Db.QueryRow(sql).Scan(&serverNr)
	if err != nil {
		logger.Logger.Println(err)
		return serverNum, err
	}
	serverNum.Number = serverNr

	return serverNum, nil
}

// CreateServer :
func CreateServer(args map[string]interface{}) (interface{}, error) {
	uuid, err := uuidgen.UUIDgen()
	if err != nil {
		logger.Logger.Println("Failed to generate uuid!")
		return nil, err
	}

	server := model.Server{
		UUID:       uuid,
		SubnetUUID: args["subnet_uuid"].(string),
		OS:         args["os"].(string),
		ServerName: args["server_name"].(string),
		ServerDesc: args["server_desc"].(string),
		CPU:        args["cpu"].(int),
		Memory:     args["memory"].(int),
		DiskSize:   args["disk_size"].(int),
		Status:     args["status"].(string),
		UserUUID:   args["user_uuid"].(string),
	}

	sql := "insert into server(uuid, subnet_uuid, os, server_name, server_desc, cpu, memory, disk_size, status, user_uuid, created_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now())"
	stmt, err := mysql.Db.Prepare(sql)
	if err != nil {
		logger.Logger.Println(err)
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

// UpdateServer :
func UpdateServer(args map[string]interface{}) (interface{}, error) {
	var err error

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
		if !subnetUUIDOk && !osOk && !serverNameOk && !serverDescOk && !cpuOk && !memoryOk && !diskSizeOk && !statusOk && !userUUIDOk {
			return nil, nil
		}
		sql := "update server set"
		if subnetUUIDOk {
			sql += " subnet_uuid = '" + server.SubnetUUID + "'"
			if osOk || serverNameOk || serverDescOk || cpuOk || memoryOk || diskSizeOk || statusOk || userUUIDOk {
				sql += ", "
			}
		}
		if osOk {
			sql += " os = '" + server.OS + "'"
			if serverNameOk || serverDescOk || cpuOk || memoryOk || diskSizeOk || statusOk || userUUIDOk {
				sql += ", "
			}
		}
		if serverNameOk {
			sql += " server_name = '" + server.ServerName + "'"
			if serverDescOk || cpuOk || memoryOk || diskSizeOk || statusOk || userUUIDOk {
				sql += ", "
			}
		}
		if serverDescOk {
			sql += " server_desc = '" + server.ServerDesc + "'"
			if cpuOk || memoryOk || diskSizeOk || statusOk || userUUIDOk {
				sql += ", "
			}
		}
		if cpuOk {
			sql += " cpu = " + strconv.Itoa(server.CPU)
			if memoryOk || diskSizeOk || statusOk || userUUIDOk {
				sql += ", "
			}
		}
		if memoryOk {
			sql += " memory = " + strconv.Itoa(server.Memory)
			if diskSizeOk || statusOk || userUUIDOk {
				sql += ", "
			}
		}
		if diskSizeOk {
			sql += " disk_size = " + strconv.Itoa(server.DiskSize)
			if statusOk || userUUIDOk {
				sql += ", "
			}
		}
		if statusOk {
			sql += " status = '" + server.Status + "'"
			if userUUIDOk {
				sql += ", "
			}
		}
		if userUUIDOk {
			sql += " user_uuid = " + server.UserUUID
		}
		sql += " where uuid = ?"

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

	return nil, err
}

// DeleteServer :
func DeleteServer(args map[string]interface{}) (interface{}, error) {
	var err error

	requestedUUID, ok := args["uuid"].(string)
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
			return nil, err
		}
		logger.Logger.Println(result.RowsAffected())

		return requestedUUID, nil
	}

	return requestedUUID, err
}
