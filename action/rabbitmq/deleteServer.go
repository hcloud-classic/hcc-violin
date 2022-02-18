package rabbitmq

import (
	dbsql "database/sql"
	"hcc/violin/action/grpc/client"
	"hcc/violin/daoext"
	"hcc/violin/lib/config"
	"hcc/violin/lib/harpUtil"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"innogrid.com/hcloud-classic/pb"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func printLogDoDeleteServerRoutineQueue(serverUUID string, msg string) {
	logger.Logger.Println("DoDeleteServerRoutineQueue(): server_uuid=" + serverUUID + ": " + msg)
}

// DoDeleteServerRoutineQueue : Do delete server stages of the queue
func DoDeleteServerRoutineQueue(routineServerUUID string, token string) {
	var routineError error

	var nodes []pb.Node

	var subnetIsInactive = false
	var subnet *pb.Subnet

	var sql string
	var stmt *dbsql.Stmt

	var err2 error

	var errCode uint64
	var errText string

	var cmd *exec.Cmd
	var harpVNUM int

	routineError = updateServerStatus(routineServerUUID, "Deleting")
	if routineError != nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"violin / update_server_status",
			"Failed",
			routineError.Error(),
			token)

		goto ERROR
	}
	_ = client.RC.WriteServerAction(
		routineServerUUID,
		"violin / update_server_status",
		"Success",
		"",
		token)

	printLogDoDeleteServerRoutineQueue(routineServerUUID, "Getting nodes list")
	nodes, routineError = client.RC.GetNodeList(routineServerUUID)
	if routineError != nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"flute / list_node",
			"Failed",
			routineError.Error(),
			token)

		goto ERROR
	}
	_ = client.RC.WriteServerAction(
		routineServerUUID,
		"flute / list_node",
		"Success",
		"",
		token)

	printLogDoDeleteServerRoutineQueue(routineServerUUID, "Getting subnet info")
	subnet, routineError = client.RC.GetSubnetByServer(routineServerUUID)
	if routineError != nil {
		if strings.Contains(routineError.Error(), "no rows in result set") {
			subnetIsInactive = true
			printLogDoDeleteServerRoutineQueue(routineServerUUID, "If seems the subnet is already changed to inactive state")
		} else {
			if routineError != nil {
				_ = client.RC.WriteServerAction(
					routineServerUUID,
					"harp / subnet",
					"Failed",
					routineError.Error(),
					token)

				goto ERROR
			}
		}
	}
	_ = client.RC.WriteServerAction(
		routineServerUUID,
		"harp / subnet",
		"Success",
		"",
		token)

	printLogDoDeleteServerRoutineQueue(routineServerUUID, "Deleting HCC Bench docker container")
	harpVNUM = harpUtil.GetHarpVNUM(subnet.Gateway)
	if routineError != nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"docker / HCC Bench",
			"Failed",
			routineError.Error(),
			token)
	}
	cmd = exec.Command("docker", "rm", "-f", "hccweb_"+strconv.Itoa(harpVNUM))
	printLogDoDeleteServerRoutineQueue(routineServerUUID, "Running docker command: "+cmd.String())
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

	if len(nodes) != 0 {
		printLogDoDeleteServerRoutineQueue(routineServerUUID, "Turning off nodes")
		routineError = daoext.DoTurnOffNodes(routineServerUUID, nodes)
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

			printLogDoDeleteServerRoutineQueue(routineServerUUID, "Wait for turning off nodes... (Remained time: "+strconv.FormatInt(i, 10)+"sec)")
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
		printLogDoDeleteServerRoutineQueue(routineServerUUID, "Deleting DHCPD configuration")
		routineError = client.RC.DeleteDHCPDConfig(subnet.UUID)
		if routineError != nil {
			logger.Logger.Println("Failed to delete DHCPD configuration")
		}
		if routineError != nil {
			_ = client.RC.WriteServerAction(
				routineServerUUID,
				"harp / delete_dhcpd_conf",
				"Failed",
				routineError.Error(),
				token)

			goto ERROR
		}
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"harp / delete_dhcpd_conf",
			"Success",
			"",
			token)
	}

	printLogDoDeleteServerRoutineQueue(routineServerUUID, "Deleting AdaptiveIP")
	_, routineError = client.RC.DeleteAdaptiveIPServer(routineServerUUID)
	if routineError != nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"harp / delete_adaptiveip_server",
			"Failed",
			routineError.Error(),
			token)

		goto ERROR
	}
	_ = client.RC.WriteServerAction(
		routineServerUUID,
		"harp / delete_adaptiveip_server",
		"Success",
		"",
		token)

	printLogDoDeleteServerRoutineQueue(routineServerUUID, "Deleting volumes")
	routineError = daoext.DoDeleteVolume(routineServerUUID)
	if routineError != nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"cello / delete_volume",
			"Failed",
			routineError.Error(),
			token)

		goto ERROR
	}
	_ = client.RC.WriteServerAction(
		routineServerUUID,
		"cello / delete_volume",
		"Success",
		"",
		token)

	printLogDoDeleteServerRoutineQueue(routineServerUUID, "Re-setting nodes info")
	for i := range nodes {
		_, routineError = client.RC.UpdateNode(&pb.ReqUpdateNode{
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
		if routineError != nil {
			_ = client.RC.WriteServerAction(
				routineServerUUID,
				"flute / update_node",
				"Failed",
				routineError.Error(),
				token)

			goto ERROR
		}
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"flute / update_node",
			"Success",
			"",
			token)
	}

	if !subnetIsInactive && subnet != nil {
		printLogDoDeleteServerRoutineQueue(routineServerUUID, "Re-setting subnet info")
		routineError = client.RC.UpdateSubnet(&pb.ReqUpdateSubnet{
			Subnet: &pb.Subnet{
				UUID:           subnet.UUID,
				ServerUUID:     "-",
				LeaderNodeUUID: "-",
				OS:             "-",
			},
		})
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
	}

	printLogDoDeleteServerRoutineQueue(routineServerUUID, "Deleting the server info from the database")
	sql = "delete from server_list where uuid = ?"
	stmt, routineError = mysql.Prepare(sql)
	if routineError != nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"delete server info",
			"Failed",
			routineError.Error(),
			token)

		goto ERROR
	}
	defer func() {
		_ = stmt.Close()
	}()
	_, err2 = stmt.Exec(routineServerUUID)
	if err2 != nil {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"delete server info",
			"Failed",
			err2.Error(),
			token)

		goto ERROR
	}
	defer func() {
		_ = stmt.Close()
	}()
	_ = client.RC.WriteServerAction(
		routineServerUUID,
		"delete server info",
		"Success",
		"",
		token)

	printLogDoDeleteServerRoutineQueue(routineServerUUID, "Deleting server nodes of the server from the database")
	_, errCode, errText = daoext.DeleteServerNodeByServerUUID(&pb.ReqDeleteServerNodeByServerUUID{
		ServerUUID: routineServerUUID,
	})
	if errCode != 0 {
		_ = client.RC.WriteServerAction(
			routineServerUUID,
			"violin / delete_server_node",
			"Failed",
			errText,
			token)

		goto ERROR
	}
	_ = client.RC.WriteServerAction(
		routineServerUUID,
		"violin / delete_server_node",
		"Success",
		"",
		token)

	_ = client.RC.WriteServerAlarm(routineServerUUID, "Delete Server", "Server has been successfully deleted.!ViolinToken!"+token)

	return

ERROR:
	printLogDoDeleteServerRoutineQueue(routineServerUUID, routineError.Error())
	err := updateServerStatus(routineServerUUID, "Failed")
	if err != nil {
		logger.Logger.Println("DoUpdateServerNodesRoutineQueue(): Failed to update server status as failed")
	}

	_ = client.RC.WriteServerAlarm(routineServerUUID,
		"Failed to delete the server",
		routineError.Error())
}
