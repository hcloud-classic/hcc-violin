package driver

import (
	"hcc/violin/data"
	"hcc/violin/http"
	"hcc/violin/lib/logger"
	"hcc/violin/model"
)

func GetSubnet(subnetUUID string) (model.Subnet, error) {
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

	var subnetData data.SubnetData

	result, err := http.DoHTTPRequest("harp", true, subnetData, query, false)
	if err != nil {
		return subnetData.Data.Subnet, err
	}

	return result.(data.SubnetData).Data.Subnet, nil
}

func UpdateSubnet(subnetUUID string, serverUUID string) (interface{}, error) {
	query := "mutation _ {\n" +
		"	update_subnet(uuid: \"" + subnetUUID + "\", server_uuid: \"" + serverUUID + "\"){\n" +
		"		uuid\n" +
		"		server_uuid\n" +
		"	}\n" +
		"}"

	var subnetData data.SubnetData

	result, err := http.DoHTTPRequest("harp", true, subnetData, query, false)
	if err != nil {
		return subnetData.Data.Subnet, err
	}

	return result.(data.SubnetData).Data.Subnet, nil
}

func CreateDHCPDConfig(subnetUUID string, nodeUUIDsStr string) error {
	query := "mutation _ {\n" +
		"	create_dhcpd_conf(subnet_uuid: \"" + subnetUUID + "\", node_uuids: \"" + nodeUUIDsStr + "\")\n" +
		"}"

	_, err := http.DoHTTPRequest("harp", false, nil, query, false)
	if err != nil {
		return err
	}

	logger.Logger.Println("CreateDHCPDConfig: Successfully created dhcpd config for subnetUUID=" + subnetUUID)

	return nil
}
