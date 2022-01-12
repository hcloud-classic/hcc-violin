package rabbitmq

import (
	"hcc/violin/action/grpc/client"
	"hcc/violin/daoext"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"net"
	"strconv"
	"strings"
	"time"

	"innogrid.com/hcloud-classic/pb"
)

func printLogDoUpdateServerRoutineQueue(serverUUID string, msg string) {
	logger.Logger.Println("DoUpdateServerNodesRoutineQueue(): server_uuid=" + serverUUID + ": " + msg)
}

// DoUpdateServerNodesRoutineQueue : Do update server stages of the queue
func DoUpdateServerNodesRoutineQueue(routineServerUUID string, routineSubnet *pb.Subnet, routineNodes []pb.Node,
	routineFirstIP net.IP, routineLastIP net.IP, token string) {
	var previousNodes []pb.Node
	var duplicatedNodeUUIDs []string

	var routineError error

	var previousNodesDetailStr string
	var newNodesDetailStr string

	printLogDoUpdateServerRoutineQueue(routineServerUUID, "Updating subnet info")
	routineError = daoext.DoUpdateSubnet(routineSubnet.UUID, routineSubnet.LeaderNodeUUID, routineServerUUID, routineSubnet.OS)
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

	printLogDoUpdateServerRoutineQueue(routineServerUUID, "Creating DHCPD config file")
	routineError = daoext.DoCreateDHCPDConfig(routineSubnet.UUID, routineServerUUID)
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

	previousNodes, routineError = client.RC.GetNodeList(routineServerUUID)
	if routineError != nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"flute / list_node (Get previous nodes)",
			"Failed",
			routineError.Error(),
			token)

		goto ERROR
	}

	for i := range routineNodes {
		skipUpdate := false

		newNodesDetailStr += "NodeName: " + routineNodes[i].NodeName + ", " +
			"UUID: " + routineNodes[i].UUID + "\\n"

		for j := range previousNodes {
			if previousNodes[j].UUID == routineNodes[i].UUID {
				skipUpdate = true
				duplicatedNodeUUIDs = append(duplicatedNodeUUIDs, previousNodes[j].UUID)
				break
			}
		}

		if skipUpdate {
			continue
		}

		_, routineError = client.RC.UpdateNode(&pb.ReqUpdateNode{
			Node: &pb.Node{
				UUID:    routineNodes[i].UUID,
				GroupID: routineSubnet.GroupID,
			},
		})
		if routineError != nil {
			_ = client.RC.WriteServerAction(
				routineServerUUID,
				"flute / update_node (New)",
				"Failed",
				routineError.Error(),
				token)

			goto ERROR
		}

		_, _, errStr := daoext.CreateServerNode(&pb.ReqCreateServerNode{
			ServerNode: &pb.ServerNode{
				ServerUUID: routineServerUUID,
				NodeUUID:   routineNodes[i].UUID,
			},
		})
		if routineError != nil {
			_ = client.RC.WriteServerAction(
				routineServerUUID,
				"violin / create_server_node (New)",
				"Failed",
				errStr,
				token)

			goto ERROR
		}
	}

	for i := range previousNodes {
		duplicated := false

		previousNodesDetailStr += "NodeName: " + previousNodes[i].NodeName + ", " +
			"UUID: " + previousNodes[i].UUID + "\\n"

		for _, nodeUUID := range duplicatedNodeUUIDs {
			if nodeUUID == previousNodes[i].UUID {
				duplicated = true
				break
			}
		}

		if duplicated {
			continue
		}

		_, routineError = client.RC.UpdateNode(&pb.ReqUpdateNode{
			Node: &pb.Node{
				UUID:    previousNodes[i].UUID,
				GroupID: int64(-1),
			},
		})
		if routineError != nil {
			_ = client.RC.WriteServerAction(
				routineServerUUID,
				"flute / update_node (Previous)",
				"Failed",
				routineError.Error(),
				token)

			goto ERROR
		}

		err := daoext.DeleteServerNodeByNodeUUID(routineNodes[i].UUID)
		if err != nil {
			_ = client.RC.WriteServerAction(
				routineServerUUID,
				"violin / delete_server_node (Previous)",
				"Failed",
				routineError.Error(),
				token)

			goto ERROR
		}
	}

	printLogDoUpdateServerRoutineQueue(routineServerUUID, "Turning off nodes")
	routineError = daoext.DoTurnOffNodes(routineServerUUID, previousNodes)
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
		isAllNodesTurnedOff := true

		printLogDoUpdateServerRoutineQueue(routineServerUUID, "Waiting for turning off nodes... (Remained time: "+strconv.FormatInt(i, 10)+"sec)")
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

	printLogDoUpdateServerRoutineQueue(routineServerUUID, "Turning on nodes")
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

	routineError = updateServerStatus(routineServerUUID, "Booting")
	if routineError != nil {
		logger.Logger.Println("DoUpdateServerNodesRoutineQueue(): Failed to update server status as booting")
	}

	printLogDoUpdateServerRoutineQueue(routineServerUUID, "Preparing controlAction")

	printLogDoUpdateServerRoutineQueue(routineServerUUID, "Running Hcc CLI")
	routineError = HccCLI(routineServerUUID, routineFirstIP, routineLastIP, token, routineSubnet.Gateway)
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

	_ = client.RC.WriteServerAlarm(routineServerUUID,
		"Server Nodes are changed",
		"[Previous Nodes]"+"\\n"+previousNodesDetailStr+"\\n"+
			"[New Nodes]"+"\\n"+newNodesDetailStr)

	return

ERROR:
	printLogDoUpdateServerRoutineQueue(routineServerUUID, routineError.Error())
	err := updateServerStatus(routineServerUUID, "Failed")
	if err != nil {
		logger.Logger.Println("DoUpdateServerNodesRoutineQueue(): Failed to update server status as failed")
	}

	_ = client.RC.WriteServerAlarm(routineServerUUID,
		"Failed to change Server Nodes",
		routineError.Error())
}
