package rabbitmq

import (
	"hcc/violin/action/grpc/client"
	pb "hcc/violin/action/grpc/pb/rpcviolin"
	"hcc/violin/daoext"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"net"
	"strconv"
	"strings"
	"time"
)

type createServerDataStruct struct {
	RoutineServerUUID string                 `json:"routine_server_uuid"`
	RoutineSubnet     pb.Subnet              `json:"routine_subnet"`
	RoutineNodes      []pb.Node              `json:"routine_nodes"`
	CelloParams       map[string]interface{} `json:"cello_params"`
	RoutineFirstIP    net.IP                 `json:"routine_first_ip"`
	RoutineLastIP     net.IP                 `json:"routine_last_ip"`
	Token             string                 `json:"token"`
}

func printLogDoCreateServerRoutineQueue(serverUUID string, msg string) {
	logger.Logger.Println("DoCreateServerRoutineQueue(): server_uuid=" + serverUUID + ": " + msg)
}

func updateServerStatus(serverUUID string, status string) error {
	sql := "update server set status = '" + status + "' where uuid = ?"

	logger.Logger.Println("UpdateServerStatus sql : ", sql)

	stmt, err := mysql.Prepare(sql)
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

// DoCreateServerRoutineQueue : Do create server stages of the queue
func DoCreateServerRoutineQueue(routineServerUUID string, routineSubnet *pb.Subnet, routineNodes []pb.Node,
	celloParams map[string]interface{}, routineFirstIP net.IP, routineLastIP net.IP, token string) {
	var routineError error

	printLogDoCreateServerRoutineQueue(routineServerUUID, "Creating os volume")
	routineError = daoext.DoCreateVolume(routineServerUUID, celloParams, "os", routineFirstIP, routineSubnet.Gateway)
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

	printLogDoCreateServerRoutineQueue(routineServerUUID, "Creating data volume")
	routineError = daoext.DoCreateVolume(routineServerUUID, celloParams, "data", routineFirstIP, routineSubnet.Gateway)
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

	printLogDoCreateServerRoutineQueue(routineServerUUID, "Updating subnet info")
	routineError = daoext.DoUpdateSubnet(routineSubnet.UUID, routineSubnet.LeaderNodeUUID, routineServerUUID)
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

	printLogDoCreateServerRoutineQueue(routineServerUUID, "Creating DHCPD config file")
	routineError = daoext.DoCreateDHCPDConfig(routineSubnet.UUID, routineServerUUID, routineNodes)
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

	printLogDoCreateServerRoutineQueue(routineServerUUID, "Turning off nodes")
	routineError = daoext.DoTurnOffNodes(routineServerUUID, routineNodes)
	if routineError != nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"flute / off_node",
			"Failed",
			routineError.Error(),
			token)

		goto ERROR
	}
	_ = client.RC.WriteServerAction(
		routineServerUUID,
		"flute / off_node",
		"Success",
		"",
		token)

	for i := config.Flute.TurnOffNodesWaitTimeSec; i >= 1; i-- {
		var isAllNodesTurnedOff = true

		printLogDoCreateServerRoutineQueue(routineServerUUID, "Waiting for turning off nodes... (Remained time: "+strconv.FormatInt(i, 10)+"sec)")
		for i := range routineNodes {
			resGetNodePowerState, _ := client.RC.GetNodePowerState(routineNodes[i].UUID)
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

	printLogDoCreateServerRoutineQueue(routineServerUUID, "Turning on nodes")
	routineError = daoext.DoTurnOnNodes(routineServerUUID, routineSubnet.LeaderNodeUUID, routineNodes)
	if routineError != nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"flute / on_node",
			"Failed",
			routineError.Error(),
			token)

		goto ERROR
	}
	_ = client.RC.WriteServerAction(
		routineServerUUID,
		"flute / on_node",
		"Success",
		"",
		token)

	printLogDoCreateServerRoutineQueue(routineServerUUID, "Preparing controlAction")

	printLogDoCreateServerRoutineQueue(routineServerUUID, "Running Hcc CLI")
	routineError = HccCLI(routineServerUUID, routineFirstIP, routineLastIP)
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

	return

ERROR:
	printLogDoCreateServerRoutineQueue(routineServerUUID, routineError.Error())
	err := updateServerStatus(routineServerUUID, "Failed")
	if err != nil {
		logger.Logger.Println("DoCreateServerRoutineQueue(): Failed to update server status as failed")
	}
}
