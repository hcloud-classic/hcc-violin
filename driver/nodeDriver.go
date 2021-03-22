package driver

import (
	"hcc/violin/data"
	"hcc/violin/http"
	"innogrid.com/hcloud-classic/pb"
	"strconv"
)

// SchedulingNodes :
func SchedulingNodes(userQuota *pb.Quota) (interface{}, error) {
	// Json Type
	query := "mutation _{\n" +
		"	schedule_nodes(server_uuid:\"" + userQuota.ServerUUID + "\", cpu:" + strconv.Itoa(int(userQuota.CPU)) + ", memory:" + strconv.Itoa(int(userQuota.Memory)) + ", nr_node:" + strconv.Itoa(int(userQuota.NumberOfNodes)) + ") {\n" +
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
