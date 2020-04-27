package driver

import (
	"hcc/violin/data"
	"hcc/violin/http"
	"hcc/violin/model"

	"strconv"
)

//  VncControl : codex
func VncControl(vncOpt model.Vnc) (interface{}, error) {
	// Json Type
	query := "mutation _{\n" +
		"	control_vnc(server_uuid:\"" + vncOpt.ServerUUID + "\", target_ip:" + vncOpt.TargetIP + ", target_port:" + vncOpt.TargetPort + ", action:" + vncOpt.ActionClassify + ") {\n" +
		" 	 node_uuid\n" +
		"	  server_uuid\n" +
		"	  target_ip\n" +
		"     target_port\n" +
		"     target_pass\n" +
		"     websocket_port\n" +
		"     vnc_info\n" +
		"     action\n" +
		"	  }" +
		"}"

	var VNCNodes data.VncNodeData
	result, err := http.DoHTTPRequest("violin_novnc", true, VNCNodes, query, true)

	if err != nil {
		return SchedulingNodes, err
	}

	return result, nil
}

func SchedulingNodes(userquota model.Quota) (interface{}, error) {
	// Json Type
	query := "mutation _{\n" +
		"	schedule_nodes(server_uuid:\"" + userquota.ServerUUID + "\", cpu:" + strconv.Itoa(userquota.CPU) + ", memory:" + strconv.Itoa(userquota.Memory) + ", nr_node:" + strconv.Itoa(userquota.NumberOfNodes) + ") {\n" +
		" 	 node_uuid\n" +
		"	  }" +
		"}"

	//String
	// query := "mutation _{\n" +
	// 	"	selected_nodes (server_uuid:\"" + userquota.ServerUUID + "\", cpu:" + strconv.Itoa(userquota.CPU) + ", memory:" + strconv.Itoa(userquota.Memory) + ", nr_node:" + strconv.Itoa(userquota.NumberOfNodes) + ") " +
	// 	"}"

	var SchedulingNodes data.ScheduledNodeData

	result, err := http.DoHTTPRequest("violin_scheduler", true, SchedulingNodes, query, false)

	if err != nil {
		return SchedulingNodes, err
	}

	return result, nil
}

// OnNode : Turn on the node by sending WOL magic packet
func OnNode(macAddr string) (interface{}, error) {
	query := "mutation _ {\n" +
		"	on_node(mac:\"" + macAddr + "\")\n" +
		"}"

	result, err := http.DoHTTPRequest("flute", false, nil, query, false)
	if err != nil {
		return "", err
	}

	return result, nil
}

// GetSingleNode : Get not activated nodes info from flute module
func GetSingleNode(NodeUUID string) (interface{}, error) {
	query := "query {\n" +
		"	node(uuid: \"" + NodeUUID + "\" ) {\n" +
		"		uuid\n" +
		"		bmc_mac_addr\n" +
		"		bmc_ip\n" +
		"		pxe_mac_addr\n" +
		"		status\n" +
		"		cpu_cores\n" +
		"		memory\n" +
		"		description\n" +
		"		created_at\n" +
		"		active\n" +
		"	}\n" +
		"}"

	var singleNodeData data.SingleNodeData

	result, err := http.DoHTTPRequest("flute", true, singleNodeData, query, true)
	if err != nil {
		return singleNodeData, err
	}
	return result, nil
}

// GetNodes : Get not activated nodes info from flute module
func GetNodes() (interface{}, error) {
	query := "query {\n" +
		"	all_node(active: 0) {\n" +
		"		uuid\n" +
		"		bmc_mac_addr\n" +
		"		bmc_ip\n" +
		"		pxe_mac_addr\n" +
		"		status\n" +
		"		cpu_cores\n" +
		"		memory\n" +
		"		description\n" +
		"		created_at\n" +
		"		active\n" +
		"	}\n" +
		"}"

	var allNodeData data.AllNodeData

	result, err := http.DoHTTPRequest("flute", true, allNodeData, query, false)
	if err != nil {
		return allNodeData, err
	}

	return result, nil
}

// UpdateNode : Add server_uuid information to each nodes
func UpdateNode(node model.Node, serverUUID string) error {
	query := "mutation{\n" +
		"	update_node(uuid:\"" + node.UUID + "\", server_uuid:\"" + serverUUID + "\"){\n" +
		"		uuid\n" +
		"	}\n" +
		"}"

	_, err := http.DoHTTPRequest("flute", false, nil, query, false)
	if err != nil {
		return err
	}

	return nil
}
