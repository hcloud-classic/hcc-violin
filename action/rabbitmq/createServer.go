package rabbitmq

import (
	"hcc/violin/action/grpc/client"
	"hcc/violin/daoext"
	"hcc/violin/lib/config"
	"hcc/violin/lib/harpUtil"
	"hcc/violin/lib/iputil"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"innogrid.com/hcloud-classic/pb"
)

type createServerDataStruct struct {
	RoutineServerUUID string                 `json:"routine_server_uuid"`
	RoutineServerOS   string                 `json:"routine_server_os"`
	RoutineSubnet     pb.Subnet              `json:"routine_subnet"`
	RoutineNodes      []pb.Node              `json:"routine_nodes"`
	CelloParams       map[string]interface{} `json:"cello_params"`
	RoutineFirstIP    net.IP                 `json:"routine_first_ip"`
	RoutineLastIP     net.IP                 `json:"routine_last_ip"`
	Token             string                 `json:"token"`
	Action            string                 `json:"action"`
}

func printLogDoCreateServerRoutineQueue(serverUUID string, msg string) {
	logger.Logger.Println("DoCreateServerRoutineQueue(): server_uuid=" + serverUUID + ": " + msg)
}

func updateServerStatus(serverUUID string, status string) error {
	sql := "update server_list set status = '" + status + "' where uuid = ?"

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
func DoCreateServerRoutineQueue(routineServerUUID string, routineServerOS string, routineSubnet *pb.Subnet, routineNodes []pb.Node,
	celloParams map[string]interface{}, routineFirstIP net.IP, routineLastIP net.IP, token string) {
	var routineError error
	var cmd *exec.Cmd
	var harpVNUM int
	var leaderNodeIP net.IP

	printLogDoCreateServerRoutineQueue(routineServerUUID, "Creating os volume")
	routineError = daoext.DoCreateVolume(routineServerUUID, celloParams, "os", routineFirstIP, routineSubnet.Gateway)
	if routineError != nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"cello / create_volume (OS)",
			"Failed",
			routineError.Error(),
			token)

		goto ErrorCreateVolumeOs
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

		goto ErrorCreateVolumeData
	}
	_ = client.RC.WriteServerAction(
		routineServerUUID,
		"cello / create_volume (Data)",
		"Success",
		"",
		token)

	printLogDoCreateServerRoutineQueue(routineServerUUID, "Updating subnet info")
	routineError = daoext.DoUpdateSubnet(routineSubnet.UUID, routineSubnet.LeaderNodeUUID, routineServerUUID, routineServerOS)
	if routineError != nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"harp / update_subnet",
			"Failed",
			routineError.Error(),
			token)

		goto ErrorUpdateSubnet
	}
	_ = client.RC.WriteServerAction(
		routineServerUUID,
		"harp / update_subnet",
		"Success",
		"",
		token)

	printLogDoCreateServerRoutineQueue(routineServerUUID, "Creating DHCPD config file")
	routineError = daoext.DoCreateDHCPDConfig(routineSubnet.UUID, routineServerUUID)
	if routineError != nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"harp / create_dhcpd_conf",
			"Failed",
			routineError.Error(),
			token)

		goto ErrorCreateDhcpConfig
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

		goto ErrorOffNode
	}
	_ = client.RC.WriteServerAction(
		routineServerUUID,
		"flute / off_node",
		"Success",
		"",
		token)

	for i := config.Flute.TurnOffNodesWaitTimeSec; i >= 1; i-- {
		isAllNodesTurnedOff := true

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

		goto ErrorOnNode
	}
	_ = client.RC.WriteServerAction(
		routineServerUUID,
		"flute / on_node",
		"Success",
		"",
		token)

	routineError = updateServerStatus(routineServerUUID, "Booting")
	if routineError != nil {
		logger.Logger.Println("DoCreateServerRoutineQueue(): Failed to update server status as booting")
	}

	printLogDoCreateServerRoutineQueue(routineServerUUID, "Preparing controlAction")

	printLogDoCreateServerRoutineQueue(routineServerUUID, "Running Hcc CLI")
	routineError = HccCLI(routineServerUUID, routineFirstIP, routineLastIP, token, routineSubnet.Gateway)
	if routineError != nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"viola / HCC_CLI",
			"Failed",
			routineError.Error(),
			token)

		goto ErrorHCCCLI
	}
	_ = client.RC.WriteServerAction(
		routineServerUUID,
		"viola / HCC_CLI",
		"Success",
		"",
		token)

	printLogDoCreateServerRoutineQueue(routineServerUUID, "Running HCC Bench docker container")
	harpVNUM = harpUtil.GetHarpVNUM(routineSubnet.Gateway)
	leaderNodeIP, _, routineError = iputil.GetFirstAndLastIPs(routineSubnet.NetworkIP, routineSubnet.Netmask)
	if routineError != nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"docker / HCC Bench",
			"Failed",
			routineError.Error(),
			token)
	}
	cmd = exec.Command("docker", "run", "-d",
		"--name", "hccweb_"+strconv.Itoa(harpVNUM), "hccweb", "-i", leaderNodeIP.String())
	printLogDoCreateServerRoutineQueue(routineServerUUID, "Running docker command: "+cmd.String())
	routineError = cmd.Run()
	if routineError != nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"docker / HCC Bench",
			"Failed",
			routineError.Error(),
			token)
	}
	if routineError == nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"docker / HCC Bench",
			"Success",
			"",
			token)
	}

	return

ErrorOnNode:
	_ = daoext.DoTurnOffNodes(routineServerUUID, routineNodes)
ErrorOffNode:
	_ = client.RC.DeleteDHCPDConfig(routineSubnet.UUID)
ErrorCreateDhcpConfig:
	_ = client.RC.UpdateSubnet(&pb.ReqUpdateSubnet{
		Subnet: &pb.Subnet{
			UUID:           routineSubnet.UUID,
			ServerUUID:     "-",
			LeaderNodeUUID: "-",
			OS:             "-",
		},
	})
ErrorUpdateSubnet:
ErrorCreateVolumeData:
	_ = daoext.DoDeleteVolume(routineServerUUID)
ErrorCreateVolumeOs:
	for i := range routineNodes {
		_, _ = client.RC.UpdateNode(&pb.ReqUpdateNode{
			Node: &pb.Node{
				UUID:       routineNodes[i].UUID,
				ServerUUID: "-",
				// gRPC use 0 value for unset. So I will use -1 for unset node_num. - ish
				NodeNum: -1,
				// gRPC use 0 value for unset. So I will use 9 value for inactive. - ish
				Active: 9,
				NodeIP: "-",
			},
		})
	}

	_, _, _ = daoext.DeleteServerNodeByServerUUID(&pb.ReqDeleteServerNodeByServerUUID{
		ServerUUID: routineServerUUID,
	})
ErrorHCCCLI:
	printLogDoCreateServerRoutineQueue(routineServerUUID, routineError.Error())
	err := updateServerStatus(routineServerUUID, "Failed")
	if err != nil {
		logger.Logger.Println("DoCreateServerRoutineQueue(): Failed to update server status as failed")
	}
}
