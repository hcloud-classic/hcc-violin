package dao

import (
	dbsql "database/sql"
	"hcc/violin/action/grpc/client"
	"hcc/violin/action/grpc/pb/rpcflute"
	"hcc/violin/action/grpc/pb/rpcharp"
	pb "hcc/violin/action/grpc/pb/rpcviolin"
	"hcc/violin/action/rabbitmq"
	hccerr "hcc/violin/lib/errors"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
)

// ReadServer : Get infos of a server
func ReadServer(uuid string) (*pb.Server, uint64, string) {
	var server pb.Server

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
	err := mysql.Db.QueryRow(sql, uuid).Scan(
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
		errStr := "ReadServer(): " + err.Error()
		logger.Logger.Println(errStr)
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, hccerr.ViolinSQLNoResult, errStr
		}
		return nil, hccerr.ViolinSQLOperationFail, errStr
	}

	server.UUID = uuid
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
		return nil, hccerr.ViolinInternalTimeStampConversionError, errStr
	}

	return &server, 0, ""
}

// ReadServerList : Get list of servers with selected infos
func ReadServerList(in *pb.ReqGetServerList) (*pb.ResGetServerList, uint64, string) {
	var serverList pb.ResGetServerList
	var servers []pb.Server
	var pservers []*pb.Server

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
		return nil, hccerr.ViolinGrpcArgumentError, "ReadServerList(): please insert row and page arguments or leave arguments as empty state"
	}

	sql := "select * from server where status != 'Deleted'"

	if in.Server != nil {
		reqServer := in.Server

		uuid = reqServer.UUID
		uuidOk := len(uuid) != 0
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
		stmt, err = mysql.Db.Query(sql, row, row*(page-1))
	} else {
		sql += " order by created_at desc"
		stmt, err = mysql.Db.Query(sql)
	}

	if err != nil {
		errStr := "ReadServerList(): " + err.Error()
		logger.Logger.Println(errStr)
		return nil, hccerr.ViolinSQLOperationFail, errStr
	}
	defer func() {
		_ = stmt.Close()
	}()

	for stmt.Next() {
		err := stmt.Scan(&uuid, &subnetUUID, &os, &serverName, &serverDesc, &cpu, &memory, &diskSize, &status, &userUUID, &createdAt)
		if err != nil {
			errStr := "ReadServerList(): " + err.Error()
			logger.Logger.Println(errStr)
			if strings.Contains(err.Error(), "no rows in result set") {
				return nil, hccerr.ViolinSQLNoResult, errStr
			}
			return nil, hccerr.ViolinSQLOperationFail, errStr
		}

		_createdAt, err := ptypes.TimestampProto(createdAt)
		if err != nil {
			errStr := "ReadServerList(): " + err.Error()
			logger.Logger.Println(errStr)
			return nil, hccerr.ViolinInternalTimeStampConversionError, errStr
		}

		servers = append(servers, pb.Server{
			UUID:       uuid,
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
	err := mysql.Db.QueryRow(sql).Scan(&serverNr)
	if err != nil {
		errStr := "ReadServerNum(): " + err.Error()
		logger.Logger.Println(errStr)
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, hccerr.ViolinSQLNoResult, errStr
		}
		return nil, hccerr.ViolinSQLOperationFail, errStr
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
	nodes, err := doGetNodes(&userQuota)
	if err != nil {
		return nil, hccerr.ViolinGrpcGetNodesError, "doGetAvailableNodes(): " + err.Error()
	}

	return nodes, 0, ""
}

func doCreateServerRoutine(server *pb.Server, nodes []pb.Node, token string) error {
	celloParams := make(map[string]interface{})
	celloParams["user_uuid"] = server.UserUUID
	celloParams["os"] = server.OS
	celloParams["disk_size"] = strconv.Itoa(int(server.DiskSize))

	logger.Logger.Println("doCreateServerRoutine(): Getting subnet info from harp module")
	serverSubnet, subnet, err := doGetSubnet(server.SubnetUUID)
	if err != nil {
		return err
	}

	logger.Logger.Println("doCreateServerRoutine(): ", serverSubnet, subnet)

	logger.Logger.Println("doCreateServerRoutine(): Getting leaderNodeUUID from first of nodes[]")
	subnet.LeaderNodeUUID = nodes[0].UUID

	logger.Logger.Println("doCreateServerRoutine(): Getting IP address range")
	firstIP, lastIP := doGetIPRange(serverSubnet, nodes)

	go func(routineServerUUID string, routineSubnet *pb.Subnet, routineNodes []pb.Node,
		celloParams map[string]interface{}, routineFirstIP net.IP, routineLastIP net.IP) {
		var routineError error

		printLogCreateServerRoutine(routineServerUUID, "Creating os volume")
		routineError = doCreateVolume(routineServerUUID, celloParams, "os", routineFirstIP, routineSubnet.Gateway)
		if routineError != nil {
			_ = client.RC.WriteServerAction(
				routineServerUUID,
				"cello / create_volume (OS)",
				"Failed",
				routineError.Error(),
				token)

			goto ERROR
		}
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"cello / create_volume (OS)",
			"Success",
			"",
			token)

		printLogCreateServerRoutine(routineServerUUID, "Creating data volume")
		routineError = doCreateVolume(routineServerUUID, celloParams, "data", routineFirstIP, routineSubnet.Gateway)
		if routineError != nil {
			_ = client.RC.WriteServerAction(
				routineServerUUID,
				"cello / create_volume (Data)",
				"Failed",
				routineError.Error(),
				token)

			goto ERROR
		}
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"cello / create_volume (Data)",
			"Success",
			"",
			token)

		printLogCreateServerRoutine(routineServerUUID, "Updating subnet info")
		routineError = doUpdateSubnet(routineSubnet.UUID, routineSubnet.LeaderNodeUUID, routineServerUUID)
		if routineError != nil {
			_ = client.RC.WriteServerAction(
				routineServerUUID,
				"harp / update_subnet",
				"Failed",
				routineError.Error(),
				token)

			goto ERROR
		}
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"harp / update_subnet",
			"Success",
			"",
			token)

		printLogCreateServerRoutine(routineServerUUID, "Creating DHCPD config file")
		routineError = doCreateDHCPDConfig(routineSubnet.UUID, routineServerUUID, routineNodes)
		if routineError != nil {
			_ = client.RC.WriteServerAction(
				routineServerUUID,
				"harp / create_dhcpd_conf",
				"Failed",
				routineError.Error(),
				token)

			goto ERROR
		}
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"harp / create_dhcpd_conf",
			"Success",
			"",
			token)

		//printLogCreateServerRoutine(routineServerUUID, "Turning off nodes")
		//routineError = doTurnOffNodes(routineServerUUID, routineNodes)
		//if routineError != nil {
		//	_ = client.RC.WriteServerAction(
		//		routineServerUUID,
		//		"flute / off_node",
		//		"Failed",
		//		routineError.Error(),
		//		token)
		//
		//	goto ERROR
		//}
		//_ = client.RC.WriteServerAction(
		//	routineServerUUID,
		//	"flute / off_node",
		//	"Success",
		//	"",
		//	token)
		//
		//printLogCreateServerRoutine(routineServerUUID, "Waiting for turning off nodes... ("+strconv.Itoa(int(config.Flute.TurnOffNodesWaitTimeSec))+"sec)")
		//time.Sleep(time.Second * time.Duration(config.Flute.TurnOffNodesWaitTimeSec))
		//
		//printLogCreateServerRoutine(routineServerUUID, "Turning on nodes")
		//routineError = doTurnOnNodes(routineServerUUID, routineSubnet.LeaderNodeUUID, routineNodes)
		//if routineError != nil {
		//	_ = client.RC.WriteServerAction(
		//		routineServerUUID,
		//		"flute / on_node",
		//		"Failed",
		//		routineError.Error(),
		//		token)
		//
		//	goto ERROR
		//}
		//_ = client.RC.WriteServerAction(
		//	routineServerUUID,
		//	"flute / on_node",
		//	"Success",
		//	"",
		//	token)

		printLogCreateServerRoutine(routineServerUUID, "Preparing controlAction")

		printLogCreateServerRoutine(routineServerUUID, "Running Hcc CLI")
		routineError = rabbitmq.HccCLI(routineServerUUID, routineFirstIP, routineLastIP)
		if routineError != nil {
			_ = client.RC.WriteServerAction(
				routineServerUUID,
				"viola / HCC_CLI",
				"Failed",
				routineError.Error(),
				token)

			goto ERROR
		}
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"viola / HCC_CLI",
			"Success",
			"",
			token)

		// while checking Cello DB cluster status is runnig in N times, until retry is expired

		return

	ERROR:
		printLogCreateServerRoutine(routineServerUUID, routineError.Error())
		err = UpdateServerStatus(routineServerUUID, "Failed")
		if err != nil {
			logger.Logger.Println("doCreateServerRoutine(): Failed to update server status as failed")
		}
	}(server.UUID, subnet, nodes, celloParams, firstIP, lastIP)

	return nil
}

func checkCreateServerArgs(reqServer *pb.Server) bool {
	subnetUUIDOk := len(reqServer.GetSubnetUUID()) != 0
	osOk := len(reqServer.GetOS()) != 0
	serverNameOk := len(reqServer.GetServerName()) != 0
	serverDescOk := len(reqServer.GetServerDesc()) != 0
	cpuOk := reqServer.GetCPU() != 0
	memoryOk := reqServer.GetMemory() != 0
	diskSizeOk := reqServer.GetDiskSize() != 0
	userUUIDOk := len(reqServer.GetUserUUID()) != 0

	return !(subnetUUIDOk && osOk && serverNameOk && serverDescOk && cpuOk && memoryOk && diskSizeOk && userUUIDOk)
}

// CreateServer : Create a server
func CreateServer(in *pb.ReqCreateServer) (*pb.Server, *hccerr.HccErrorStack) {
	var cpuCores int32 = 0
	var memory int32 = 0
	var serverUUID string
	var nodes []pb.Node
	var server pb.Server

	var sql string
	var stmt *dbsql.Stmt

	var err error
	var errCode uint64
	var errStr string
	errStack := hccerr.NewHccErrorStack()

	reqServer := in.GetServer()
	if reqServer == nil {
		errStack.Push(&hccerr.HccError{ErrCode: hccerr.ViolinGrpcArgumentError, ErrText: "CreateServer(): Server is nil"})

		goto ERROR
	}

	logger.Logger.Println("CreateServer(): Generating server UUID")
	serverUUID, err = doGenerateServerUUID()
	if err != nil {
		errStack.Push(&hccerr.HccError{ErrCode: hccerr.ViolinInternalUUIDGenerationError, ErrText: "CreateServer(): " + err.Error()})

		goto ERROR
	}

	if checkCreateServerArgs(reqServer) {
		errStack.Push(&hccerr.HccError{ErrCode: hccerr.ViolinGrpcArgumentError, ErrText: "CreateServer(): some of arguments are missing"})

		goto ERROR
	}
	//Scheduler
	nodes, errCode, errStr = doGetAvailableNodes(in, serverUUID)
	if errCode != 0 {
		errStack.Push(&hccerr.HccError{ErrCode: errCode, ErrText: errStr})
		errStack.Push(&hccerr.HccError{ErrCode: hccerr.ViolinInternalGetAvailableNodesError, ErrText: "CreateServer(): Failed to get available nodes"})

		goto ERROR
	}

	for i := range nodes {
		cpuCores += nodes[i].CPUCores
		memory += nodes[i].Memory
	}

	server = pb.Server{
		UUID:       serverUUID,
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

	sql = "insert into server(uuid, subnet_uuid, os, server_name, server_desc, cpu, memory, disk_size, status, user_uuid, created_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now())"
	stmt, err = mysql.Db.Prepare(sql)
	if err != nil {
		errStr := "CreateServer(): " + err.Error()
		logger.Logger.Println(errStr)
		errStack.Push(&hccerr.HccError{ErrCode: hccerr.ViolinSQLOperationFail, ErrText: errStr})

		goto ERROR
	}
	defer func() {
		_ = stmt.Close()
	}()
	_, err = stmt.Exec(server.UUID, server.SubnetUUID, server.OS, server.ServerName, server.ServerDesc, server.CPU, server.Memory, server.DiskSize, server.Status, server.UserUUID)
	if err != nil {
		errStr := "CreateServer(): " + err.Error()
		logger.Logger.Println(errStr)
		errStack.Push(&hccerr.HccError{ErrCode: hccerr.ViolinSQLOperationFail, ErrText: errStr})

		goto ERROR
	}

	err = doCreateServerRoutine(&server, nodes, in.GetToken())
	if err != nil {
		errStack.Push(&hccerr.HccError{ErrCode: hccerr.ViolinInternalCreateServerRoutineError, ErrText: err.Error()})

		goto ERROR
	}

	return &server, errStack.ConvertReportForm()
ERROR:
	errStack.Push(&hccerr.HccError{
		ErrCode: hccerr.ViolinInternalCreateServerFailed,
		ErrText: "CreateServer(): Failed to create server",
	})

	return nil, errStack.ConvertReportForm()
}

func checkUpdateServerArgs(reqServer *pb.Server) bool {
	subnetUUIDOk := len(reqServer.SubnetUUID) != 0
	osOk := len(reqServer.OS) != 0
	serverNameOk := len(reqServer.ServerName) != 0
	serverDescOk := len(reqServer.ServerDesc) != 0
	cpuOk := reqServer.CPU != 0
	memoryOk := reqServer.Memory != 0
	diskSizeOk := reqServer.DiskSize != 0
	userUUIDOk := len(reqServer.UserUUID) != 0

	return !subnetUUIDOk && !osOk && !serverNameOk && !serverDescOk && !cpuOk && !memoryOk && !diskSizeOk && !userUUIDOk
}

// UpdateServer : Update infos of the server
func UpdateServer(in *pb.ReqUpdateServer) (*pb.Server, uint64, string) {
	// TODO : Update server stages
	// TODO : Currently UpdateServer() only updates infos of the server. Need some works to call other modules.

	if in.Server == nil {
		return nil, hccerr.ViolinGrpcArgumentError, "UpdateServer(): server is nil"
	}
	reqServer := in.Server

	requestedUUID := reqServer.GetUUID()
	requestedUUIDOk := len(requestedUUID) != 0
	if !requestedUUIDOk {
		return nil, hccerr.ViolinGrpcArgumentError, "UpdateServer(): need a uuid argument"
	}

	if checkUpdateServerArgs(reqServer) {
		return nil, hccerr.ViolinGrpcArgumentError, "UpdateServer(): need some arguments"
	}

	var subnetUUID string
	var os string
	var serverName string
	var serverDesc string
	var cpu int
	var memory int
	var diskSize int
	var status string
	var userUUID string

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

	server := new(pb.Server)
	server.UUID = requestedUUID
	server.SubnetUUID = subnetUUID
	server.OS = os
	server.ServerName = serverName
	server.ServerDesc = serverDesc
	server.CPU = int32(cpu)
	server.Memory = int32(memory)
	server.DiskSize = int32(diskSize)
	server.Status = status
	server.UserUUID = userUUID

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
		updateSet += " cpu = " + strconv.Itoa(int(server.CPU)) + ", "
	}
	if memoryOk {
		updateSet += " memory = " + strconv.Itoa(int(server.Memory)) + ", "
	}
	if diskSizeOk {
		updateSet += " disk_size = " + strconv.Itoa(int(server.DiskSize)) + ", "
	}
	if statusOk {
		updateSet += " status = '" + server.Status + "', "
	}
	if userUUIDOk {
		updateSet += " user_uuid = '" + server.UserUUID + "', "
	}

	sql += updateSet[0:len(updateSet)-2] + " where uuid = ?"

	logger.Logger.Println("update_server sql : ", sql)

	stmt, err := mysql.Db.Prepare(sql)
	if err != nil {
		errStr := "UpdateServer(): " + err.Error()
		logger.Logger.Println(errStr)
		return nil, hccerr.ViolinSQLOperationFail, errStr
	}
	defer func() {
		_ = stmt.Close()
	}()

	_, err2 := stmt.Exec(server.UUID)
	if err2 != nil {
		errStr := "UpdateServer(): " + err2.Error()
		logger.Logger.Println(errStr)
		return nil, hccerr.ViolinSQLOperationFail, errStr
	}

	server, errCode, errStr := ReadServer(server.UUID)
	if errCode != 0 {
		logger.Logger.Println("UpdateServer(): " + errStr)
	}

	return server, 0, ""
}

// UpdateServerStatus : Update status of the server
func UpdateServerStatus(serverUUID string, status string) error {
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

	_, err2 := stmt.Exec(serverUUID)
	if err2 != nil {
		logger.Logger.Println(err2)
		return err
	}

	return nil
}

// DeleteServer : Delete a server by UUID
func DeleteServer(in *pb.ReqDeleteServer) (*pb.Server, uint64, string) {
	var err error

	requestedUUID := in.GetUUID()
	requestedUUIDOk := len(requestedUUID) != 0
	if !requestedUUIDOk {
		return nil, hccerr.ViolinGrpcArgumentError, "DeleteServer(): Need a uuid argument"
	}

	server, errCode, errText := ReadServer(requestedUUID)
	if errCode != 0 {
		return nil, hccerr.ViolinGrpcRequestError, "DeleteServer(): " + errText
	}

	logger.Logger.Println("DeleteServer(): Deleting the server (ServerUUID: " + requestedUUID + ")")

	logger.Logger.Println("DeleteServer(): Getting nodes list (ServerUUID: " + requestedUUID + ")")
	nodes, err := client.RC.GetNodeList(requestedUUID)
	if err != nil {
		return nil, hccerr.ViolinGrpcRequestError, "DeleteServer(): Failed to get nodes (" + err.Error() + ")"
	}

	logger.Logger.Println("DeleteServer(): Getting subnet info (ServerUUID: " + requestedUUID + ")")
	subnet, err := client.RC.GetSubnetByServer(requestedUUID)
	if err != nil {
		return nil, hccerr.ViolinGrpcRequestError, "DeleteServer(): Failed to get subnet info (" + err.Error() + ")"
	}

	//logger.Logger.Println("DeleteServer(): Turning off nodes (ServerUUID: " + requestedUUID + ")")
	//err = doTurnOffNodes(requestedUUID, nodes)
	//if err != nil {
	//	return nil, hccerr.ViolinGrpcRequestError, "DeleteServer(): Failed to turning off nodes (" + err.Error() + ")"
	//}

	logger.Logger.Println("DeleteServer(): Deleting DHCPD configuration (ServerUUID: " + requestedUUID + ")")
	err = client.RC.DeleteDHCPDConfig(subnet.UUID)
	if err != nil {
		return nil, hccerr.ViolinGrpcRequestError, "DeleteServer(): Failed to delete DHCPD configuration (" + err.Error() + ")"
	}

	logger.Logger.Println("DeleteServer(): Re-setting subnet info (ServerUUID: " + requestedUUID + ")")
	err = client.RC.UpdateSubnet(&rpcharp.ReqUpdateSubnet{
		Subnet: &pb.Subnet{
			UUID:           subnet.UUID,
			ServerUUID:     "-",
			LeaderNodeUUID: "-",
		},
	})
	if err != nil {
		return nil, hccerr.ViolinGrpcRequestError, "DeleteServer(): Failed to re-setting subnet info (" + err.Error() + ")"
	}

	logger.Logger.Println("DeleteServer(): Deleting AdaptiveIP (ServerUUID: " + requestedUUID + ")")
	_, err = client.RC.DeleteAdaptiveIPServer(requestedUUID)
	if err != nil {
		return nil, hccerr.ViolinGrpcRequestError, "DeleteServer(): Failed to delete AdaptiveIP (" + err.Error() + ")"
	}

	// TODO : Delete volumes of the server

	logger.Logger.Println("DeleteServer(): Deleting AdaptiveIP (ServerUUID: " + requestedUUID + ")")

	err = doDeleteVolume(requestedUUID)
	if err != nil {
		return nil, hccerr.ViolinGrpcRequestError, "DeleteServer(): Failed to delete Volume (" + err.Error() + ")"
	}

	logger.Logger.Println("DeleteServer(): Re-setting nodes info (ServerUUID: " + requestedUUID + ")")
	for i := range nodes {
		_, err = client.RC.UpdateNode(&rpcflute.ReqUpdateNode{
			Node: &pb.Node{
				UUID:       nodes[i].UUID,
				ServerUUID: "",
				// gRPC use 0 value for unset. So I will use 9 value for inactive. - ish
				Active: 9,
			},
		})
		if err != nil {
			return nil, hccerr.ViolinGrpcRequestError, "DeleteServer(): Failed to re-setting nodes info (" + err.Error() + ")"
		}
	}

	logger.Logger.Println("DeleteServer(): Deleting the server info from the database (UUID: " + requestedUUID + ")")
	sql := "delete from server where uuid = ?"
	stmt, err := mysql.Db.Prepare(sql)
	if err != nil {
		errStr := "DeleteServer(): Failed to deleting the server info from the database (" + err.Error() + ")"
		logger.Logger.Println(errStr)
		return nil, hccerr.ViolinSQLOperationFail, errStr
	}
	defer func() {
		_ = stmt.Close()
	}()
	_, err2 := stmt.Exec(requestedUUID)
	if err2 != nil {
		errStr := "DeleteServer(): Failed to deleting the server info from the database (" + err2.Error() + ")"
		logger.Logger.Println(errStr)
		return nil, hccerr.ViolinSQLOperationFail, errStr
	}

	return server, 0, ""
}
