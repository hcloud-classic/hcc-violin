package graphql

import "hcc/violin/model"

// CreateDisk : Create os or data disk
func CreateDisk(volume model.Volume, serverUUID string) error {
	query := "mutation _ {\n" +
		"	create_volume(size:0, filesystem:\"" + volume.Filesystem + "\", server_uuid:\"" + serverUUID + "\", use_type:\"" + volume.UseType + "\", user_uuid:\"" + volume.UseType + "\", network_ip:\"" + volume.NetworkIP + "\") {\n" +
		"		uuid\n" +
		"		size\n" +
		"		filesystem\n" +
		"		server_uuid\n" +
		"		use_type\n" +
		"		user_uuid\n" +
		"		created_at\n" +
		"	}\n" +
		"}"

	_, err := DoHTTPRequest("cello", false, nil, query)
	if err != nil {
		return err
	}

	return nil
}
