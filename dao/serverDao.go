package dao

import (
	dbsql "database/sql"
	"errors"
	"github.com/golang/protobuf/ptypes"
	pb "hcc/violin/action/grpc/rpcviolin"
	"hcc/violin/action/rabbitmq"
	cmdutil "hcc/violin/lib/cmdUtil"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"net"
	"strconv"
	"time"
)

// ReadServer : Get infos of a server
func ReadServer(uuid string) (*pb.Server, error) {
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
		logger.Logger.Println(err)
		return nil, err
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
		logger.Logger.Println(err)
		return nil, err
	}

	return &server, nil
}

// ReadServerList : Get list of servers with selected infos
func ReadServerList(in *pb.ReqGetServerList) (*pb.ResGetServerList, error) {
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
		return nil, errors.New("please insert row and page arguments or leave arguments as empty state")
	}

	sql := "select * from server where status != 'Deleted'"

	if in.Server != nil {
		reqServer := in.Server

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
		logger.Logger.Println(err.Error())
		return nil, err
	}
	defer func() {
		_ = stmt.Close()
	}()

	for stmt.Next() {
		err := stmt.Scan(&uuid, &subnetUUID, &os, &serverName, &serverDesc, &cpu, &memory, &diskSize, &status, &userUUID, &createdAt)
		if err != nil {
			logger.Logger.Println(err.Error())
			return nil, err
		}

		_createdAt, err := ptypes.TimestampProto(createdAt)
		if err != nil {
			logger.Logger.Println(err)
			return nil, err
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

	return &serverList, nil
}

// ReadServerNum : Get the number of servers
func ReadServerNum() (*pb.ResGetServerNum, error) {
	var serverNum pb.ResGetServerNum
	var serverNr int64

	sql := "select count(*) from server where status != 'Deleted'"
	err := mysql.Db.QueryRow(sql).Scan(&serverNr)
	if err != nil {
		logger.Logger.Println(err)
		return nil, err
	}
	serverNum.Num = serverNr

	return &serverNum, nil
}

func doGetAvailableNodes(in *pb.ReqCreateServer) ([]pb.Node, error) {
	var nodes []pb.Node
	server := in.GetServer()

	var userQuota pb.Quota
	userQuota.ServerUUID = server.UUID
	userQuota.CPU = server.CPU
	userQuota.Memory = server.Memory
	userQuota.NumberOfNodes = in.GetNrNode()

	logger.Logger.Println("createServer: Getting available nodes from flute module")
	nodes, err := doGetNodes(&userQuota)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func doCreateServerRoutine(server *pb.Server, nodes []pb.Node) error {
	celloParams := make(map[string]interface{})
	celloParams["user_uuid"] = server.UserUUID
	celloParams["os"] = server.OS
	celloParams["disk_size"] = strconv.Itoa(int(server.DiskSize))

	logger.Logger.Println("createServer: Getting subnet info from harp module")
	serverSubnet, subnet, err := doGetSubnet(server.SubnetUUID)
	if err != nil {
		return err
	}
	logger.Logger.Println("createServer: ", serverSubnet, subnet)

	logger.Logger.Println("createServer: Getting leaderNodeUUID from first of nodes[]")
	subnet.LeaderNodeUUID = nodes[0].UUID

	logger.Logger.Println("createServer: Getting IP address range")
	firstIP, lastIP := doGetIPRange(serverSubnet, nodes)

	go func(routineServerUUID string, routineSubnet *pb.Subnet, routineNodes []pb.Node,
		celloParams map[string]interface{}, routineFirstIP net.IP, routineLastIP net.IP) {
		var routineError error

		printLogCreateServerRoutine(routineServerUUID, "Creating os volume")
		routineError = doCreateVolume(routineServerUUID, celloParams, "os", routineFirstIP, routineSubnet.Gateway)
		if routineError != nil {
			goto ERROR
		}

		printLogCreateServerRoutine(routineServerUUID, "Creating data volume")
		routineError = doCreateVolume(routineServerUUID, celloParams, "data", routineFirstIP, routineSubnet.Gateway)
		if routineError != nil {
			goto ERROR
		}

		printLogCreateServerRoutine(routineServerUUID, "Updating subnet info")
		routineError = doUpdateSubnet(routineSubnet.UUID, routineSubnet.LeaderNodeUUID, routineServerUUID)
		if routineError != nil {
			goto ERROR
		}

		printLogCreateServerRoutine(routineServerUUID, "Creating DHCPD config file")
		routineError = doCreateDHCPDConfig(routineSubnet.UUID, routineServerUUID, routineNodes)
		if routineError != nil {
			goto ERROR
		}

		printLogCreateServerRoutine(routineServerUUID, "Turning on nodes")
		routineError = doTurnOnNodes(routineServerUUID, routineSubnet.LeaderNodeUUID, routineNodes)
		if routineError != nil {
			goto ERROR
		}

		printLogCreateServerRoutine(routineServerUUID, "Preparing controlAction")

		printLogCreateServerRoutine(routineServerUUID, "Running Hcc CLI")
		routineError = rabbitmq.HccCLI(routineServerUUID, routineFirstIP, routineLastIP)
		if routineError != nil {
			goto ERROR
		}
		// while checking Cello DB cluster status is runnig in N times, until retry is expired

		return

	ERROR:
		printLogCreateServerRoutine(routineServerUUID, routineError.Error())
		err = UpdateServerStatus(routineServerUUID, "Failed")
		if err != nil {
			logger.Logger.Println("createServerRoutine: Failed to update server status as failed")
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
	statusOk := len(reqServer.GetStatus()) != 0
	userUUIDOk := len(reqServer.GetUserUUID()) != 0

	return !(subnetUUIDOk && osOk && serverNameOk && serverDescOk && cpuOk && memoryOk && diskSizeOk && statusOk && userUUIDOk)
}

// CreateServer : Create a server
func CreateServer(in *pb.ReqCreateServer) (*pb.Server, error) {
	reqServer := in.GetServer()
	if reqServer == nil {
		return nil, errors.New("server is nil")
	}

	logger.Logger.Println("createServer: Generating server UUID")
	serverUUID, err := doGenerateServerUUID()
	if err != nil {
		return nil, err
	}

	if checkCreateServerArgs(reqServer) {
		return nil, errors.New("some of arguments are missing")
	}

	nodes, err := doGetAvailableNodes(in)
	if err != nil {
		return nil, err
	}

	var cpuCores int32 = 0
	var memory int32 = 0
	for i := range nodes {
		cpuCores += nodes[i].CPUCores
		memory += nodes[i].Memory
	}

	server := pb.Server{
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

	err = doCreateServerRoutine(&server, nodes)
	if err != nil {
		return nil, err
	}

	return &server, nil
}

func checkUpdateServerArgs(reqServer *pb.Server) bool {
	subnetUUIDOk := len(reqServer.SubnetUUID) != 0
	osOk := len(reqServer.OS) != 0
	serverNameOk := len(reqServer.ServerName) != 0
	serverDescOk := len(reqServer.ServerDesc) != 0
	cpuOk := reqServer.CPU != 0
	memoryOk := reqServer.Memory != 0
	diskSizeOk := reqServer.DiskSize != 0
	statusOk := len(reqServer.Status) != 0
	userUUIDOk := len(reqServer.UserUUID) != 0

	return !subnetUUIDOk && !osOk && !serverNameOk && !serverDescOk && !cpuOk && !memoryOk && !diskSizeOk && !statusOk && !userUUIDOk
}

// UpdateServer : Update infos of the server
func UpdateServer(in *pb.ReqUpdateServer) (*pb.Server, error) {
	// TODO : Update server stages
	// TODO : Currently UpdateServer() only updates infos of the server. Need some works to call other modules.

	if in.Server == nil {
		return nil, errors.New("server is nil")
	}
	reqServer := in.Server

	requestedUUID := reqServer.GetUUID()
	requestedUUIDOk := len(requestedUUID) != 0
	if !requestedUUIDOk {
		return nil, errors.New("need a uuid argument")
	}

	if checkUpdateServerArgs(reqServer) {
		return nil, errors.New("need some arguments")
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

	server, err = ReadServer(server.UUID)
	if err != nil {
		logger.Logger.Println(err)
	}

	return server, nil
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

	result, err2 := stmt.Exec(serverUUID)
	if err2 != nil {
		logger.Logger.Println(err2)
		return err
	}
	logger.Logger.Println(result.LastInsertId())
	return nil
}

// DeleteServer : Delete a server by UUID
func DeleteServer(in *pb.ReqDeleteServer) (string, error) {
	// TODO : Delete server stages
	_ = cmdutil.RunScript("/root/script/prepare_create_server.sh")

	var err error

	requestedUUID := in.GetUUID()
	requestedUUIDOk := len(requestedUUID) != 0
	if !requestedUUIDOk {
		return "", errors.New("need a uuid argument")
	}

	sql := "delete from server where uuid = ?"
	stmt, err := mysql.Db.Prepare(sql)
	if err != nil {
		logger.Logger.Println(err.Error())
		return "", err
	}
	defer func() {
		_ = stmt.Close()
	}()
	result, err2 := stmt.Exec(requestedUUID)
	if err2 != nil {
		logger.Logger.Println(err2)
		return "", err2
	}
	logger.Logger.Println(result.RowsAffected())

	return requestedUUID, nil
}

//func TestServer(params graphql.ResolveParams) (interface{}, error) {
//	var userQuota model.Quota
//	userQuota.ServerUUID = "COdex"
//	logger.Logger.Println("$$$$$$$$$$$$")
//
//	userQuota.CPU = 0
//	userQuota.Memory = 0
//	userQuota.NumberOfNodes = 2
//	nodes, err := NodeScheduler(userQuota)
//	if err != nil {
//		return nil, err
//	}
//
//	return nodes, nil
//}
