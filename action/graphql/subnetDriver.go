package graphql

import (
	"hcc/violin/lib/logger"
	"reflect"
)

// GetSubnet : Get subnet info from harp module
func GetSubnet(subnetUUID string) (SubnetData, error) {
	query := "query {\n" +
		"	subnet(uuid:\"" + subnetUUID + "\"){\n" +
		"		uuid\n" +
		"		network_ip\n" +
		"		netmask\n" +
		"		gateway\n" +
		"		next_server\n" +
		"		name_server\n" +
		"		domain_name\n" +
		"		server_uuid\n" +
		"		leader_node_uuid\n" +
		"		os\n" +
		"		subnet_name\n" +
		"		created_at\n" +
		"	}\n" +
		"}"

	var subnetData SubnetData

	result, err := DoHTTPRequest(true, reflect.ValueOf(subnetData).Interface(), query)
	if err != nil {
		return subnetData, err
	}

	return result.(SubnetData), nil
}

// UpdateSubnet : Add server_uuid to subnet
func UpdateSubnet(subnetUUID string, serverUUID string) (SubnetData, error) {
	query := "mutation _ {\n" +
		"	update_subnet(uuid: \"" + subnetUUID + "\", server_uuid: \"" + serverUUID + "\"){\n" +
		"		uuid\n" +
		"		server_uuid\n" +
		"	}\n" +
		"}"

	var subnetData SubnetData

	result, err := DoHTTPRequest(true, reflect.ValueOf(subnetData).Interface(), query)
	if err != nil {
		return subnetData, err
	}

	return result.(SubnetData), nil
}

// CreateDHCPDConfig : Add server_uuid to subnet
func CreateDHCPDConfig(subnetUUID string, nodeUUIDsStr string) error {
	query := "mutation _ {\n" +
		"	create_dhcpd_conf(subnet_uuid: \"" + subnetUUID + "\", node_uuids: \"" + nodeUUIDsStr + "\")\n" +
		"}"

	_, err := DoHTTPRequest(false, nil, query)
	if err != nil {
		return err
	}

	logger.Logger.Println("CreateDHCPDConfig: Successfully created dhcpd config for subnetUUID=" + subnetUUID)

	return nil
}
