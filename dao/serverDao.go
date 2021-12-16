package dao

import (
	dbsql "database/sql"
	"errors"
	"fmt"
	"hcc/violin/action/grpc/client"
	"hcc/violin/action/rabbitmq"
	"hcc/violin/daoext"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"innogrid.com/hcloud-classic/hcc_errors"
	"innogrid.com/hcloud-classic/pb"
)

// ReadServer : Get infos of a server
func ReadServer(uuid string) (*pb.Server, uint64, string) {
	var server pb.Server

	var groupID int64
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
	row := mysql.Db.QueryRow(sql, uuid)
	err := mysql.QueryRowScan(row,
		&uuid,
		&groupID,
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
		errStr := "ReadServer(): " + err.Error()
		// logger.Logger.Println(errStr)
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, hcc_errors.ViolinSQLNoResult, errStr
		}
		return nil, hcc_errors.ViolinSQLOperationFail, errStr
	}

	server.UUID = uuid
	server.GroupID = groupID
	server.SubnetUUID = subnetUUID
	server.OS = os
	server.ServerName = serverName
	server.ServerDesc = serverDesc
	server.CPU = int32(cpu)
	server.Memory = int32(memory)
	server.DiskSize = int32(diskSize)
	server.Status = status
	server.UserUUID = userUUID
	server.CreatedAt = timestamppb.New(createdAt)

	return &server, 0, ""
}

// ReadServerList : Get list of servers with selected infos
func ReadServerList(in *pb.ReqGetServerList) (*pb.ResGetServerList, uint64, string) {
	var serverList pb.ResGetServerList
	var servers []pb.Server
	var pservers []*pb.Server

	var uuid string
	var groupID int64
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

	var isLimit bool
	row := in.GetRow()
	rowOk := row != 0
	page := in.GetPage()
	pageOk := page != 0
	if !rowOk && !pageOk {
		isLimit = false
	} else if rowOk && pageOk {
		isLimit = true
	} else {
		return nil, hcc_errors.ViolinGrpcArgumentError, "ReadServerList(): please insert row and page arguments or leave arguments as empty state"
	}

	sql := "select * from server where status != 'Deleted'"

	if in.Server != nil {
		reqServer := in.Server

		uuid = reqServer.UUID
		uuidOk := len(uuid) != 0
		groupID = reqServer.GroupID
		groupIDOk := groupID != 0
		subnetUUID = reqServer.SubnetUUID
		subnetUUIDOk := len(subnetUUID) != 0
		os = reqServer.OS
		osOk := len(os) != 0
		serverName = reqServer.ServerName
		serverNameOk := len(serverName) != 0
		serverDesc = reqServer.ServerDesc
		serverDescOk := len(serverDesc) != 0
		cpu = int(reqServer.CPU)
		cpuOk := cpu != 0
		memory = int(reqServer.Memory)
		memoryOk := memory != 0
		diskSize = int(reqServer.DiskSize)
		diskSizeOk := diskSize != 0
		status = reqServer.Status
		statusOk := len(status) != 0
		userUUID = reqServer.UserUUID
		userUUIDOk := len(userUUID) != 0

		if uuidOk {
			sql += " and uuid like '%" + uuid + "%'"
		}
		if groupIDOk {
			sql += " and group_id = " + strconv.Itoa(int(groupID))
		}
		if subnetUUIDOk {
			sql += " and subnet_uuid like '%" + subnetUUID + "%'"
		}
		if osOk {
			sql += " and os like '%" + os + "%'"
		}
		if serverNameOk {
			sql += " and server_name like '%" + serverName + "%'"
		}
		if serverDescOk {
			sql += " and server_desc like '%" + serverDesc + "%'"
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
			sql += " and status like '%" + status + "%'"
		}
		if userUUIDOk {
			sql += " and user_uuid like '%" + userUUID + "%'"
		}
	}

	var stmt *dbsql.Rows
	var err error
	if isLimit {
		sql += " order by created_at desc limit ? offset ?"
		stmt, err = mysql.Query(sql, row, row*(page-1))
	} else {
		sql += " order by created_at desc"
		stmt, err = mysql.Query(sql)
	}

	if err != nil {
		errStr := "ReadServerList(): " + err.Error()
		logger.Logger.Println(errStr)
		return nil, hcc_errors.ViolinSQLOperationFail, errStr
	}
	defer func() {
		_ = stmt.Close()
	}()

	for stmt.Next() {
		err := stmt.Scan(&uuid, &groupID, &subnetUUID, &os, &serverName, &serverDesc, &cpu, &memory, &diskSize, &status, &userUUID, &createdAt)
		if err != nil {
			errStr := "ReadServerList(): " + err.Error()
			logger.Logger.Println(errStr)
			if strings.Contains(err.Error(), "no rows in result set") {
				return nil, hcc_errors.ViolinSQLNoResult, errStr
			}
			return nil, hcc_errors.ViolinSQLOperationFail, errStr
		}

		servers = append(servers, pb.Server{
			UUID:       uuid,
			GroupID:    groupID,
			SubnetUUID: subnetUUID,
			OS:         os,
			ServerName: serverName,
			ServerDesc: serverDesc,
			CPU:        int32(cpu),
			Memory:     int32(memory),
			DiskSize:   int32(diskSize),
			Status:     status,
			UserUUID:   userUUID,
			CreatedAt:  timestamppb.New(createdAt),
		})
	}

	for i := range servers {
		pservers = append(pservers, &servers[i])
	}

	serverList.Server = pservers

	return &serverList, 0, ""
}

// ReadServerNum : Get the number of servers
func ReadServerNum(in *pb.ReqGetServerNum) (*pb.ResGetServerNum, uint64, string) {
	var serverNum pb.ResGetServerNum
	var serverNr int64
	groupID := in.GetGroupID()

	if groupID == 0 {
		return nil, hcc_errors.ViolinGrpcArgumentError, "ReadServerNum(): please insert a group_id argument"
	}

	sql := "select count(*) from server where status != 'Deleted' and group_id = " + strconv.Itoa(int(groupID))
	row := mysql.Db.QueryRow(sql)
	err := mysql.QueryRowScan(row, &serverNr)
	if err != nil {
		errStr := "ReadServerNum(): " + err.Error()
		logger.Logger.Println(errStr)
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, hcc_errors.ViolinSQLNoResult, errStr
		}
		return nil, hcc_errors.ViolinSQLOperationFail, errStr
	}
	serverNum.Num = serverNr

	return &serverNum, 0, ""
}

func doGetAvailableNodes(in *pb.ReqCreateServer, UUID string) ([]pb.Node, uint64, string) {
	var nodes []pb.Node
	server := in.GetServer()

	var userQuota pb.Quota
	userQuota.ServerUUID = UUID
	userQuota.CPU = server.CPU
	userQuota.Memory = server.Memory
	userQuota.NumberOfNodes = in.GetNrNode()

	logger.Logger.Println("doGetAvailableNodes(): Getting available nodes from flute module ")
	allNodes, err := daoext.DoGetNodes(&userQuota)
	if err != nil {
		return nil, hcc_errors.ViolinGrpcGetNodesError, "doGetAvailableNodes(): " + err.Error()
	}
	fmt.Println("####\nallNodes\n#### => ", allNodes)
	for i := range allNodes {
		if server.GroupID != allNodes[i].GroupID {
			continue
		}
		nodes = append(nodes, pb.Node{
			UUID:     allNodes[i].UUID,
			CPUCores: allNodes[i].CPUCores,
			Memory:   allNodes[i].Memory,
		})
	}

	if len(nodes) == 0 {
		return nil, hcc_errors.ViolinGrpcGetNodesError, "doGetAvailableNodes(): " + "Nodes are not available from your group."
	}

	resGetQuota, errStack := client.RC.GetQuota(server.GroupID)
	if errStack != nil {
		return nil, hcc_errors.ViolinGrpcRequestError, "doGetAvailableNodes(): " + errStack.Pop().Text()
	}

	if len(nodes) > int(resGetQuota.Quota.LimitNodeCnt) {
		return nil, hcc_errors.ViolinGrpcRequestError, "doGetAvailableNodes(): Node count quota exceeded"
	}

	return nodes, 0, ""
}

func doCreateServerRoutine(server *pb.Server, nodes []pb.Node, token string) error {
	celloParams := make(map[string]interface{})
	celloParams["user_uuid"] = server.UserUUID
	celloParams["os"] = server.OS
	celloParams["disk_size"] = strconv.Itoa(int(server.DiskSize))
	celloParams["group_id"] = server.GroupID
	logger.Logger.Println("doCreateServerRoutine(): Getting subnet info from harp module")
	serverSubnet, subnet, err := daoext.DoGetSubnet(server.SubnetUUID, false)
	if err != nil {
		return err
	}

	logger.Logger.Println("doCreateServerRoutine(): ", serverSubnet, subnet)

	logger.Logger.Println("doCreateServerRoutine(): Getting leaderNodeUUID from first of nodes[]")
	subnet.LeaderNodeUUID = nodes[0].UUID

	logger.Logger.Println("doCreateServerRoutine(): Getting IP address range")
	firstIP, lastIP := daoext.DoGetIPRange(serverSubnet, nodes)

	err = rabbitmq.QueueCreateServer(server.UUID, server.OS, subnet, nodes, celloParams, firstIP, lastIP, token)
	if err != nil {
		return err
	}

	return nil
}

func doUpdateServerNodesRoutine(server *pb.Server, nodes []pb.Node, token string) error {
	logger.Logger.Println("doUpdateServerNodesRoutine(): Getting subnet info from harp module")
	serverSubnet, subnet, err := daoext.DoGetSubnet(server.SubnetUUID, true)
	if err != nil {
		return err
	}

	logger.Logger.Println("doUpdateServerNodesRoutine(): ", serverSubnet, subnet)

	logger.Logger.Println("doUpdateServerNodesRoutine(): Getting leaderNodeUUID from first of nodes[]")
	subnet.LeaderNodeUUID = nodes[0].UUID

	logger.Logger.Println("doUpdateServerNodesRoutine(): Getting IP address range")
	firstIP, lastIP := daoext.DoGetIPRange(serverSubnet, nodes)

	err = rabbitmq.QueueUpdateServerNodes(server.UUID, subnet, nodes, firstIP, lastIP, token)
	if err != nil {
		return err
	}

	return nil
}

func checkGroupIDExist(groupID int64) error {
	resGetGroupList, hccErrStack := client.RC.GetGroupList(&pb.Empty{})
	if hccErrStack != nil {
		return hccErrStack.Pop().ToError()
	}

	for _, pGroup := range resGetGroupList.Group {
		if pGroup.Id == groupID {
			return nil
		}
	}

	return errors.New("given group ID is not in the database")
}

func checkCreateServerArgs(reqServer *pb.Server) bool {
	groupIDOk := reqServer.GroupID != 0
	subnetUUIDOk := len(reqServer.GetSubnetUUID()) != 0
	osOk := len(reqServer.GetOS()) != 0
	serverNameOk := len(reqServer.GetServerName()) != 0
	serverDescOk := len(reqServer.GetServerDesc()) != 0
	cpuOk := reqServer.GetCPU() != 0
	memoryOk := reqServer.GetMemory() != 0
	diskSizeOk := reqServer.GetDiskSize() != 0
	userUUIDOk := len(reqServer.GetUserUUID()) != 0

	return !(groupIDOk && subnetUUIDOk && osOk && serverNameOk && serverDescOk && cpuOk && memoryOk && diskSizeOk && userUUIDOk)
}

// CreateServer : Create a server
func CreateServer(in *pb.ReqCreateServer) (*pb.Server, *hcc_errors.HccErrorStack) {
	var cpuCores int32
	var memory int32
	var serverUUID string
	var nodes []pb.Node
	var server pb.Server

	var sql string
	var stmt *dbsql.Stmt

	var err error
	var errCode uint64
	var errStr string
	errStack := hcc_errors.NewHccErrorStack()
	reqServer := in.GetServer()
	if reqServer == nil {
		errCode = hcc_errors.ViolinGrpcArgumentError
		errStr = "CreateServer(): Server is nil"

		goto ERROR
	}

	logger.Logger.Println("CreateServer(): Generating server UUID")
	serverUUID, err = daoext.DoGenerateServerUUID()
	if err != nil {
		errCode = hcc_errors.ViolinInternalUUIDGenerationError
		errStr = "CreateServer(): " + err.Error()

		goto ERROR
	}

	_ = client.RC.WriteServerAlarm(server.UUID, "Create Server", "Creating the server.")

	if checkCreateServerArgs(reqServer) {
		errCode = hcc_errors.ViolinGrpcArgumentError
		errStr = "CreateServer(): some of arguments are missing"

		goto ERROR
	}

	err = checkGroupIDExist(reqServer.GroupID)
	if err != nil {
		errCode = hcc_errors.ViolinInternalCreateServerRoutineError
		errStr = "CreateServer(): " + err.Error()

		goto ERROR
	}

	// Scheduler
	nodes, errCode, errStr = doGetAvailableNodes(in, serverUUID)
	if errCode != 0 {
		errCode = hcc_errors.ViolinInternalGetAvailableNodesError
		errStr = "CreateServer(): Failed to get available nodes"

		goto ERROR
	}

	for i := range nodes {
		cpuCores += nodes[i].CPUCores
		memory += nodes[i].Memory
	}

	server = pb.Server{
		UUID:       serverUUID,
		GroupID:    reqServer.GetGroupID(),
		SubnetUUID: reqServer.GetSubnetUUID(),
		OS:         reqServer.GetOS(),
		ServerName: reqServer.GetServerName(),
		ServerDesc: reqServer.GetServerDesc(),
		Status:     "Creating",
		UserUUID:   reqServer.GetUserUUID(),
		DiskSize:   reqServer.GetDiskSize(),
	}

	sql = "insert into server_list(uuid, group_id,subnet_uuid, os, server_name, server_desc, status, user_uuid, created_at) values (?, ?, ?, ?, ?, ?, ?, ?, now())"
	stmt, err = mysql.Prepare(sql)
	if err != nil {
		errCode = hcc_errors.ViolinSQLOperationFail
		errStr = "CreateServer(): " + err.Error()
		logger.Logger.Println(errStr)

		goto ERROR
	}
	defer func() {
		_ = stmt.Close()
	}()
	_, err = stmt.Exec(server.UUID, server.GroupID, server.SubnetUUID, server.OS, server.ServerName, server.ServerDesc, server.Status, server.UserUUID)
	if err != nil {
		errCode = hcc_errors.ViolinSQLOperationFail
		errStr = "CreateServer(): " + err.Error()
		logger.Logger.Println(errStr)

		goto ERROR
	}

	err = doCreateServerRoutine(&server, nodes, in.GetToken())
	if err != nil {
		errCode = hcc_errors.ViolinInternalCreateServerRoutineError
		errStr = "CreateServer(): " + err.Error()

		goto ERROR
	}

	return &server, errStack
ERROR:
	logger.Logger.Println("CreateServer(): Failed to create server")
	logger.Logger.Println("CreateServer(): errStr: ", errStr)

	_ = errStack.Push(hcc_errors.NewHccError(errCode, errStr))
	_ = client.RC.WriteServerAlarm(server.UUID, "Create Server", errStr)

	return nil, errStack
}

// ScaleUpServer : Scale up the server
func ScaleUpServer(in *pb.ReqScaleUpServer) (*pb.Server, *hcc_errors.HccErrorStack) {
	var availableNode []pb.Node
	var serverNodes []pb.Node
	var server *pb.Server

	var err error
	var errCode uint64
	var errStr string
	errStack := hcc_errors.NewHccErrorStack()

	serverUUID := in.GetServerUUID()
	if len(serverUUID) == 0 {
		errCode = hcc_errors.ViolinGrpcArgumentError
		errStr = "ScaleUpServer(): Need a serverUUID argument"

		goto ERROR
	}

	_ = client.RC.WriteServerAlarm(serverUUID, "Auto Scale Queued", "Scaling up the server.")

	server, errCode, errStr = ReadServer(serverUUID)
	if errCode != 0 {
		errCode = hcc_errors.ViolinInternalOperationFail
		errStr = "ScaleUpServer(): " + errStr

		goto ERROR
	}

	serverNodes, err = client.RC.GetNodeList(serverUUID)
	if err != nil {
		errCode = hcc_errors.ViolinInternalCreateServerRoutineError
		errStr = "ScaleUpServer(): " + err.Error()

		goto ERROR
	}
	if len(serverNodes) == 0 {
		errCode = hcc_errors.ViolinInternalCreateServerRoutineError
		errStr = "ScaleUpServer(): Failed to get server nodes"

		goto ERROR
	}

	// Scheduler
	availableNode, err = daoext.DoGetNodes(&pb.Quota{
		ServerUUID:    serverUUID,
		CPU:           serverNodes[0].CPUCores,
		Memory:        serverNodes[0].Memory,
		NumberOfNodes: 1,
	})
	if err != nil {
		errCode = hcc_errors.ViolinInternalGetAvailableNodesError
		errStr = "ScaleUpServer(): " + err.Error()
		logger.Logger.Println(errStr)

		goto ERROR
	}
	if len(availableNode) == 0 {
		errCode = hcc_errors.ViolinInternalCreateServerRoutineError
		errStr = "ScaleUpServer(): Failed to get a available node"

		goto ERROR
	}

	serverNodes = append(serverNodes, pb.Node{
		UUID:       availableNode[0].UUID,
		ServerUUID: serverUUID,
	})

	err = doUpdateServerNodesRoutine(server, serverNodes, in.GetToken())
	if err != nil {
		errCode = hcc_errors.ViolinInternalCreateServerRoutineError
		errStr = "ScaleUpServer(): " + err.Error()

		goto ERROR
	}

	server, errCode, errStr = ReadServer(serverUUID)
	if errCode != 0 {
		logger.Logger.Println("ScaleUpServer(): " + errStr)
	}

	return server, errStack
ERROR:
	logger.Logger.Println("ScaleUpServer(): Failed to scale up the server")
	logger.Logger.Println("ScaleUpServer(): errStr: ", errStr)

	_ = errStack.Push(hcc_errors.NewHccError(errCode, errStr))
	_ = client.RC.WriteServerAlarm(server.UUID, "ScaleUp Server", errStr)

	return nil, errStack
}

func checkUpdateServerArgs(reqServer *pb.Server) bool {
	serverNameOk := len(reqServer.ServerName) != 0
	serverDescOk := len(reqServer.ServerDesc) != 0
	statusOk := len(reqServer.Status) != 0

	return !serverNameOk && !serverDescOk && !statusOk
}

// UpdateServer : Update infos of the server
func UpdateServer(in *pb.ReqUpdateServer) (*pb.Server, *hcc_errors.HccErrorStack) {
	var server *pb.Server
	var reqServer *pb.Server

	var serverName string
	var serverNameOk bool
	var serverDesc string
	var serverDescOk bool
	var status string
	var statusOk bool
	var requestedUUID string
	var requestedUUIDOk bool

	var sql string
	var stmt *dbsql.Stmt
	updateSet := ""

	var err error
	var err2 error
	var errCode uint64
	var errStr string
	errStack := hcc_errors.NewHccErrorStack()

	if in.Server == nil {
		errCode = hcc_errors.ViolinGrpcArgumentError
		errStr = "UpdateServer(): server is nil"

		goto ERROR
	}
	reqServer = in.Server

	requestedUUID = reqServer.GetUUID()
	requestedUUIDOk = len(requestedUUID) != 0
	if !requestedUUIDOk {
		errCode = hcc_errors.ViolinGrpcArgumentError
		errStr = "UpdateServer(): need a uuid argument"

		goto ERROR
	}

	serverName = reqServer.ServerName
	serverNameOk = len(reqServer.ServerName) != 0
	serverDesc = reqServer.ServerDesc
	serverDescOk = len(reqServer.ServerDesc) != 0
	status = reqServer.Status
	statusOk = len(reqServer.Status) != 0

	if checkUpdateServerArgs(reqServer) {
		errCode = hcc_errors.ViolinGrpcArgumentError
		errStr = "UpdateServer(): need some arguments"

		goto ERROR
	}

	sql = "update server_list set"
	if serverNameOk {
		updateSet += " server_name = '" + serverName + "', "
	}
	if serverDescOk {
		updateSet += " server_desc = '" + serverDesc + "', "
	}
	if statusOk {
		updateSet += " status = '" + status + "', "
	}

	sql += updateSet[0:len(updateSet)-2] + " where uuid = ?"

	stmt, err = mysql.Prepare(sql)
	if err != nil {
		errCode = hcc_errors.ViolinSQLOperationFail
		errStr = "UpdateServer(): " + err.Error()
		logger.Logger.Println(errStr)

		goto ERROR
	}
	defer func() {
		_ = stmt.Close()
	}()

	_, err2 = stmt.Exec(requestedUUID)
	if err2 != nil {
		errCode = hcc_errors.ViolinSQLOperationFail
		errStr = "UpdateServer(): " + err2.Error()
		logger.Logger.Println(errStr)

		goto ERROR
	}

	server, errCode, errStr = ReadServer(requestedUUID)
	if errCode != 0 {
		logger.Logger.Println("UpdateServer(): " + errStr)
	}

	return server, errStack
ERROR:
	logger.Logger.Println("UpdateServer(): Failed to update server")
	logger.Logger.Println("UpdateServer(): errStr: ", errStr)

	_ = errStack.Push(hcc_errors.NewHccError(errCode, errStr))
	_ = client.RC.WriteServerAlarm(server.UUID, "Update Server", errStr)

	return nil, errStack
}

// DeleteServer : Delete a server by UUID
func DeleteServer(in *pb.ReqDeleteServer) (*pb.Server, uint64, string) {
	var err error

	requestedUUID := in.GetServerUUID()
	requestedUUIDOk := len(requestedUUID) != 0
	if !requestedUUIDOk {
		return nil, hcc_errors.ViolinGrpcArgumentError, "DeleteServer(): Need a uuid argument"
	}

	_ = client.RC.WriteServerAlarm(requestedUUID, "Delete Server", "Deleting the server.")

	server, errCode, errText := ReadServer(requestedUUID)
	if errCode != 0 {
		return nil, hcc_errors.ViolinGrpcRequestError, "DeleteServer(): " + errText
	}

	logger.Logger.Println("DeleteServer(): Deleting the server (ServerUUID: " + requestedUUID + ")")

	err = rabbitmq.QueueDeleteServer(requestedUUID, in.GetToken())
	if err != nil {
		return nil, hcc_errors.ViolinGrpcRequestError, "DeleteServer(): " + err.Error()
	}

	return server, 0, ""
}

// UpdateServerNodes : Update nodes of the server
func UpdateServerNodes(in *pb.ReqUpdateServerNodes) (*pb.Server, *hcc_errors.HccErrorStack) {
	var server *pb.Server

	var requestedUUID string
	var selectedNodes string
	var nodes []pb.Node
	var splitSelectedNodes []string

	var err error
	var errCode uint64
	var errStr string
	errStack := hcc_errors.NewHccErrorStack()

	requestedUUID = in.GetServerUUID()
	if len(requestedUUID) == 0 {
		errCode = hcc_errors.ViolinGrpcArgumentError
		errStr = "UpdateServerNodes(): Need a server_uuid argument"

		goto ERROR
	}

	server, errCode, errStr = ReadServer(requestedUUID)
	if errCode != 0 {
		logger.Logger.Println("UpdateServerNodes(): " + errStr)
	}

	_ = client.RC.WriteServerAlarm(server.UUID, "Update Server Nodes", "Updating the server nodes.")

	selectedNodes = in.GetSelectedNodes()
	if len(selectedNodes) == 0 {
		errCode = hcc_errors.ViolinGrpcArgumentError
		errStr = "UpdateServerNodes(): Nodes are not selected"

		goto ERROR
	}

	splitSelectedNodes = strings.Split(selectedNodes, ",")
	for _, nodeUUID := range splitSelectedNodes {
		if nodeUUID == "" {
			continue
		}

		nodes = append(nodes, pb.Node{
			UUID: nodeUUID,
		})
	}

	err = doUpdateServerNodesRoutine(server, nodes, in.GetToken())
	if err != nil {
		errCode = hcc_errors.ViolinInternalCreateServerRoutineError
		errStr = "UpdateServerNodes(): " + err.Error()

		goto ERROR
	}

	server, errCode, errStr = ReadServer(requestedUUID)
	if errCode != 0 {
		logger.Logger.Println("UpdateServerNodes(): " + errStr)
	}

	return server, errStack
ERROR:
	logger.Logger.Println("UpdateServerNodes(): Failed to update nodes of the server")
	logger.Logger.Println("UpdateServerNodes(): errStr: ", errStr)

	_ = errStack.Push(hcc_errors.NewHccError(errCode, errStr))
	_ = client.RC.WriteServerAlarm(server.UUID, "Update Server Nodes", errStr)

	return nil, errStack
}
