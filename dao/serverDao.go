package dao

import (
	dbsql "database/sql"
	"errors"
	"hcc/violin/action/grpc/client"
	"hcc/violin/action/rabbitmq"
	"hcc/violin/daoext"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"hcc/violin/model"
	"innogrid.com/hcloud-classic/hcc_errors"
	"innogrid.com/hcloud-classic/pb"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
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
		logger.Logger.Println(errStr)
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

	server.CreatedAt, err = ptypes.TimestampProto(createdAt)
	if err != nil {
		errStr := "ReadServer(): " + err.Error()
		logger.Logger.Println(errStr)
		return nil, hcc_errors.ViolinInternalTimeStampConversionError, errStr
	}

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
			sql += " and uuid = '" + uuid + "'"
		}
		if groupIDOk {
			sql += " and group_id = " + strconv.Itoa(int(groupID))
		}
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

		_createdAt, err := ptypes.TimestampProto(createdAt)
		if err != nil {
			errStr := "ReadServerList(): " + err.Error()
			logger.Logger.Println(errStr)
			return nil, hcc_errors.ViolinInternalTimeStampConversionError, errStr
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
			CreatedAt:  _createdAt})
	}

	for i := range servers {
		pservers = append(pservers, &servers[i])
	}

	serverList.Server = pservers

	return &serverList, 0, ""
}

// ReadServerNum : Get the number of servers
func ReadServerNum() (*pb.ResGetServerNum, uint64, string) {
	var serverNum pb.ResGetServerNum
	var serverNr int64

	sql := "select count(*) from server where status != 'Deleted'"
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
	nodes, err := daoext.DoGetNodes(&userQuota)
	if err != nil {
		return nil, hcc_errors.ViolinGrpcGetNodesError, "doGetAvailableNodes(): " + err.Error()
	}

	return nodes, 0, ""
}

func doCreateServerRoutine(server *pb.Server, nodes []pb.Node, token string) error {
	celloParams := make(map[string]interface{})
	celloParams["user_uuid"] = server.UserUUID
	celloParams["os"] = server.OS
	celloParams["disk_size"] = strconv.Itoa(int(server.DiskSize))

	logger.Logger.Println("doCreateServerRoutine(): Getting subnet info from harp module")
	serverSubnet, subnet, err := daoext.DoGetSubnet(server.SubnetUUID)
	if err != nil {
		return err
	}

	logger.Logger.Println("doCreateServerRoutine(): ", serverSubnet, subnet)

	logger.Logger.Println("doCreateServerRoutine(): Getting leaderNodeUUID from first of nodes[]")
	subnet.LeaderNodeUUID = nodes[0].UUID

	logger.Logger.Println("doCreateServerRoutine(): Getting IP address range")
	firstIP, lastIP := daoext.DoGetIPRange(serverSubnet, nodes)

	err = rabbitmq.QueueCreateServer(server.UUID, subnet, nodes, celloParams, firstIP, lastIP, token)
	if err != nil {
		return err
	}

	return nil
}

func checkGroupIDExist(groupID int64) error {
	resGetGroupList, hccErrStack := client.RC.GetGroupList(&pb.Empty{})
	if hccErrStack != nil {
		return (*hccErrStack.Stack())[0].ToError()
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
	var cpuCores int32 = 0
	var memory int32 = 0
	var serverUUID string
	var nodes []pb.Node
	var server pb.Server

	var totalDiskSize int32

	var sql string
	var stmt *dbsql.Stmt

	var err error
	var errCode uint64
	var errStr string
	errStack := hcc_errors.NewHccErrorStack()

	reqServer := in.GetServer()
	if reqServer == nil {
		_ = errStack.Push(hcc_errors.NewHccError(hcc_errors.ViolinGrpcArgumentError, "CreateServer(): Server is nil"))

		goto ERROR
	}

	logger.Logger.Println("CreateServer(): Generating server UUID")
	serverUUID, err = daoext.DoGenerateServerUUID()
	if err != nil {
		_ = errStack.Push(hcc_errors.NewHccError(hcc_errors.ViolinInternalUUIDGenerationError, "CreateServer(): "+err.Error()))

		goto ERROR
	}

	if checkCreateServerArgs(reqServer) {
		_ = errStack.Push(hcc_errors.NewHccError(hcc_errors.ViolinGrpcArgumentError, "CreateServer(): some of arguments are missing"))

		goto ERROR
	}

	err = checkGroupIDExist(reqServer.GroupID)
	if err != nil {
		errStr = "CreateServer(): " + err.Error()
		_ = errStack.Push(hcc_errors.NewHccError(hcc_errors.ViolinInternalCreateServerRoutineError, errStr))

		goto ERROR
	}

	//Scheduler
	nodes, errCode, errStr = doGetAvailableNodes(in, serverUUID)
	if errCode != 0 {
		_ = errStack.Push(hcc_errors.NewHccError(errCode, errStr))
		_ = errStack.Push(hcc_errors.NewHccError(hcc_errors.ViolinInternalGetAvailableNodesError, "CreateServer(): Failed to get available nodes"))

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
		CPU:        cpuCores,
		Memory:     memory,
		DiskSize:   reqServer.GetDiskSize(),
		Status:     "Creating",
		UserUUID:   reqServer.GetUserUUID(),
	}

	totalDiskSize = int32(model.OSDiskSize) + reqServer.GetDiskSize()

	sql = "insert into server(uuid, subnet_uuid, os, server_name, server_desc, cpu, memory, disk_size, status, user_uuid, created_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now())"
	stmt, err = mysql.Prepare(sql)
	if err != nil {
		errStr := "CreateServer(): " + err.Error()
		logger.Logger.Println(errStr)
		_ = errStack.Push(hcc_errors.NewHccError(hcc_errors.ViolinSQLOperationFail, errStr))

		goto ERROR
	}
	defer func() {
		_ = stmt.Close()
	}()
	_, err = stmt.Exec(server.UUID, server.SubnetUUID, server.OS, server.ServerName, server.ServerDesc, server.CPU, server.Memory, totalDiskSize, server.Status, server.UserUUID)
	if err != nil {
		errStr := "CreateServer(): " + err.Error()
		logger.Logger.Println(errStr)
		_ = errStack.Push(hcc_errors.NewHccError(hcc_errors.ViolinSQLOperationFail, errStr))

		goto ERROR
	}

	err = doCreateServerRoutine(&server, nodes, in.GetToken())
	if err != nil {
		_ = errStack.Push(hcc_errors.NewHccError(hcc_errors.ViolinInternalCreateServerRoutineError, err.Error()))

		goto ERROR
	}

	return &server, errStack.ConvertReportForm()
ERROR:
	logger.Logger.Println("CreateServer(): Failed to create server")
	logger.Logger.Println("CreateServer(): errStack: ", errStack)

	_ = errStack.Push(hcc_errors.NewHccError(
		hcc_errors.ViolinInternalCreateServerFailed,
		"CreateServer(): Failed to create server",
	))

	return nil, errStack.ConvertReportForm()
}

func checkUpdateServerArgs(reqServer *pb.Server) bool {
	groupIDOk := reqServer.GroupID != 0
	subnetUUIDOk := len(reqServer.SubnetUUID) != 0
	osOk := len(reqServer.OS) != 0
	serverNameOk := len(reqServer.ServerName) != 0
	serverDescOk := len(reqServer.ServerDesc) != 0
	cpuOk := reqServer.CPU != 0
	memoryOk := reqServer.Memory != 0
	diskSizeOk := reqServer.DiskSize != 0
	userUUIDOk := len(reqServer.UserUUID) != 0

	return !groupIDOk && !subnetUUIDOk && !osOk && !serverNameOk && !serverDescOk && !cpuOk && !memoryOk && !diskSizeOk && !userUUIDOk
}

// UpdateServer : Update infos of the server
func UpdateServer(in *pb.ReqUpdateServer) (*pb.Server, uint64, string) {
	// TODO : Update server stages
	// TODO : Currently UpdateServer() only updates infos of the server. Need some works to call other modules.

	if in.Server == nil {
		return nil, hcc_errors.ViolinGrpcArgumentError, "UpdateServer(): server is nil"
	}
	reqServer := in.Server

	requestedUUID := reqServer.GetUUID()
	requestedUUIDOk := len(requestedUUID) != 0
	if !requestedUUIDOk {
		return nil, hcc_errors.ViolinGrpcArgumentError, "UpdateServer(): need a uuid argument"
	}

	if checkUpdateServerArgs(reqServer) {
		return nil, hcc_errors.ViolinGrpcArgumentError, "UpdateServer(): need some arguments"
	}

	err := checkGroupIDExist(reqServer.GroupID)
	if err != nil {
		errStr := "UpdateServer(): " + err.Error()

		return nil, hcc_errors.ViolinGrpcArgumentError, errStr
	}

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

	sql := "update server set"
	var updateSet = ""
	if groupIDOk {
		updateSet += " group_id = " + strconv.Itoa(int(groupID)) + ", "
	}
	if subnetUUIDOk {
		updateSet += " subnet_uuid = '" + subnetUUID + "', "
	}
	if osOk {
		updateSet += " os = '" + os + "', "
	}
	if serverNameOk {
		updateSet += " server_name = '" + serverName + "', "
	}
	if serverDescOk {
		updateSet += " server_desc = '" + serverDesc + "', "
	}
	if cpuOk {
		updateSet += " cpu = " + strconv.Itoa(int(cpu)) + ", "
	}
	if memoryOk {
		updateSet += " memory = " + strconv.Itoa(int(memory)) + ", "
	}
	if diskSizeOk {
		updateSet += " disk_size = " + strconv.Itoa(int(diskSize)) + ", "
	}
	if statusOk {
		updateSet += " status = '" + status + "', "
	}
	if userUUIDOk {
		updateSet += " user_uuid = '" + userUUID + "', "
	}

	sql += updateSet[0:len(updateSet)-2] + " where uuid = ?"

	logger.Logger.Println("update_server sql : ", sql)

	stmt, err := mysql.Prepare(sql)
	if err != nil {
		errStr := "UpdateServer(): " + err.Error()
		logger.Logger.Println(errStr)
		return nil, hcc_errors.ViolinSQLOperationFail, errStr
	}
	defer func() {
		_ = stmt.Close()
	}()

	_, err2 := stmt.Exec(requestedUUID)
	if err2 != nil {
		errStr := "UpdateServer(): " + err2.Error()
		logger.Logger.Println(errStr)
		return nil, hcc_errors.ViolinSQLOperationFail, errStr
	}

	server, errCode, errStr := ReadServer(requestedUUID)
	if errCode != 0 {
		logger.Logger.Println("UpdateServer(): " + errStr)
	}

	return server, 0, ""
}

// DeleteServer : Delete a server by UUID
func DeleteServer(in *pb.ReqDeleteServer) (*pb.Server, uint64, string) {
	var err error

	requestedUUID := in.GetUUID()
	requestedUUIDOk := len(requestedUUID) != 0
	if !requestedUUIDOk {
		return nil, hcc_errors.ViolinGrpcArgumentError, "DeleteServer(): Need a uuid argument"
	}

	server, errCode, errText := ReadServer(requestedUUID)
	if errCode != 0 {
		return nil, hcc_errors.ViolinGrpcRequestError, "DeleteServer(): " + errText
	}

	logger.Logger.Println("DeleteServer(): Deleting the server (ServerUUID: " + requestedUUID + ")")

	logger.Logger.Println("DeleteServer(): Getting nodes list (ServerUUID: " + requestedUUID + ")")
	nodes, err := client.RC.GetNodeList(requestedUUID)
	if err != nil {
		return nil, hcc_errors.ViolinGrpcRequestError, "DeleteServer(): Failed to get nodes (" + err.Error() + ")"
	}

	if len(nodes) == 0 {
		logger.Logger.Println("DeleteServer(): If seems nodes are already changed to inactive state (ServerUUID: " + requestedUUID + ")")
	}

	var subnetIsInactive = false
	var subnet *pb.Subnet

	logger.Logger.Println("DeleteServer(): Getting subnet info (ServerUUID: " + requestedUUID + ")")
	subnet, err = client.RC.GetSubnetByServer(requestedUUID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			subnetIsInactive = true
			logger.Logger.Println("DeleteServer(): If seems the subnet is already changed to inactive state (ServerUUID: " + requestedUUID + ")")
		} else {
			return nil, hcc_errors.ViolinGrpcRequestError, "DeleteServer(): Failed to get subnet info (" + err.Error() + ")"
		}
	}

	if len(nodes) != 0 {
		logger.Logger.Println("DeleteServer(): Turning off nodes (ServerUUID: " + requestedUUID + ")")
		err = daoext.DoTurnOffNodes(requestedUUID, nodes)
		if err != nil {
			logger.Logger.Println("DeleteServer(): Failed to turning off nodes (Error: " + err.Error() + ", ServerUUID: " + requestedUUID + ")")
		}

		for i := config.Flute.TurnOffNodesWaitTimeSec; i >= 1; i-- {
			var isAllNodesTurnedOff = true

			logger.Logger.Println("DeleteServer(): Wait for turning off nodes... (Remained time: " + strconv.FormatInt(i, 10) + "sec, ServerUUID: " + requestedUUID + ")")
			for i := range nodes {
				resGetNodePowerState, _ := client.RC.GetNodePowerState(nodes[i].UUID)
				if strings.ToLower(resGetNodePowerState.Result) == "on" {
					isAllNodesTurnedOff = false
					break
				}
			}

			if isAllNodesTurnedOff {
				break
			}

			time.Sleep(time.Second * time.Duration(1))
		}
	}

	if !subnetIsInactive && subnet != nil {
		logger.Logger.Println("DeleteServer(): Deleting DHCPD configuration (ServerUUID: " + requestedUUID + ")")
		err = client.RC.DeleteDHCPDConfig(subnet.UUID)
		if err != nil {
			logger.Logger.Println("DeleteServer(): Failed to delete DHCPD configuration (Error: " + err.Error() + ", ServerUUID: " + requestedUUID + ")")
		}

		logger.Logger.Println("DeleteServer(): Re-setting subnet info (ServerUUID: " + requestedUUID + ")")
		err = client.RC.UpdateSubnet(&pb.ReqUpdateSubnet{
			Subnet: &pb.Subnet{
				UUID:           subnet.UUID,
				ServerUUID:     "-",
				LeaderNodeUUID: "-",
			},
		})
		if err != nil {
			logger.Logger.Println("DeleteServer(): Failed to re-setting subnet info (" + err.Error() + ")")
		}
	}

	logger.Logger.Println("DeleteServer(): Deleting AdaptiveIP (ServerUUID: " + requestedUUID + ")")
	_, err = client.RC.DeleteAdaptiveIPServer(requestedUUID)
	if err != nil {
		logger.Logger.Println("DeleteServer(): Failed to delete AdaptiveIP  (Error: " + err.Error() + ", ServerUUID: " + requestedUUID + ")")
	}

	logger.Logger.Println("DeleteServer(): Deleting volumes (ServerUUID: " + requestedUUID + ")")
	err = daoext.DoDeleteVolume(requestedUUID)
	if err != nil {
		logger.Logger.Println("DeleteServer(): Failed to delete volumes  (Error: " + err.Error() + ", ServerUUID: " + requestedUUID + ")")
	}

	logger.Logger.Println("DeleteServer(): Re-setting nodes info (ServerUUID: " + requestedUUID + ")")
	for i := range nodes {
		_, err = client.RC.UpdateNode(&pb.ReqUpdateNode{
			Node: &pb.Node{
				UUID:       nodes[i].UUID,
				ServerUUID: "-",
				// gRPC use 0 value for unset. So I will use -1 for unset node_num. - ish
				NodeNum: -1,
				// gRPC use 0 value for unset. So I will use 9 value for inactive. - ish
				Active: 9,
				NodeIP: "-",
			},
		})
		if err != nil {

			logger.Logger.Println("DeleteServer(): Failed to re-setting nodes info  (Error: " + err.Error() + ", ServerUUID: " + requestedUUID + ")")
		}
	}

	logger.Logger.Println("DeleteServer(): Deleting the server info from the database (UUID: " + requestedUUID + ")")
	sql := "delete from server where uuid = ?"
	stmt, err := mysql.Prepare(sql)
	if err != nil {
		errStr := "DeleteServer(): Failed to deleting the server info from the database  (Error: " + err.Error() + ", ServerUUID: " + requestedUUID + ")"
		logger.Logger.Println(errStr)
		return nil, hcc_errors.ViolinSQLOperationFail, errStr
	}
	defer func() {
		_ = stmt.Close()
	}()
	_, err2 := stmt.Exec(requestedUUID)
	if err2 != nil {
		errStr := "DeleteServer(): Failed to deleting the server info from the database  (Error: " + err2.Error() + ", ServerUUID: " + requestedUUID + ")"
		logger.Logger.Println(errStr)
		return nil, hcc_errors.ViolinSQLOperationFail, errStr
	}

	logger.Logger.Println("DeleteServer(): Deleting server nodes of the server from the database (ServerUUID: " + requestedUUID + ")")
	_, errCode, errText = DeleteServerNodeByServerUUID(&pb.ReqDeleteServerNodeByServerUUID{
		ServerUUID: requestedUUID,
	})
	if errCode != 0 {
		errStr := "DeleteServer(): Failed to deleting the server nodes of the server from the database  (Error: " + errText + ", ServerUUID: " + requestedUUID + ")"
		logger.Logger.Println(errStr)
		return nil, hcc_errors.ViolinSQLOperationFail, errStr
	}

	return server, 0, ""
}
