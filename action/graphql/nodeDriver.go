package graphql

import (
	"hcc/violin/model"
	"reflect"
)

// OnNode : Turn on the node by sending WOL magic packet
func OnNode(macAddr string) (string, error) {
	query := "mutation _ {\n" +
		"	on_node(mac:\"" + macAddr + "\")\n" +
		"}"

	result, err := DoHTTPRequest(false, nil, query)
	if err != nil {
		return "", err
	}

	return result.(string), nil
}

// GetNodes : Get not activated nodes info from flute module
func GetNodes() (ListNodeData, error) {
	query := "query {\n" +
		"	list_node(active: 0, row:10, page:1) {\n" +
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

	var listNodeData ListNodeData

	result, err := DoHTTPRequest(true, reflect.ValueOf(listNodeData).Interface(), query)
	if err != nil {
		return listNodeData, err
	}

	return result.(ListNodeData), nil
}

// UpdateNode : Add server_uuid information to each nodes
func UpdateNode(node model.Node, serverUUID string) error {
	query := "mutation{\n" +
		"	update_node(uuid:\"" + node.UUID + "\", server_uuid:\"" + serverUUID + "\", active: 1){\n" +
		"		uuid\n" +
		"	}\n" +
		"}"

	_, err := DoHTTPRequest(false, nil, query)
	if err != nil {
		return err
	}

	return nil
}
