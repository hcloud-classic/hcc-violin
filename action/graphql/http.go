package graphql

import (
	"encoding/json"
	"errors"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"hcc/violin/model"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Flute

// ToFluteOnNode : Turn on the node by sending WOL magic packet
func ToFluteOnNode(macAddr string) (string, error) {
	query := "mutation _ {\n" +
		"	on_node(mac:\"" + macAddr + "\")\n" +
		"}"

	client := &http.Client{Timeout: time.Duration(config.Flute.RequestTimeoutMs) * time.Millisecond}
	req, err := http.NewRequest("GET", "http://"+config.Flute.ServerAddress+":"+strconv.Itoa(int(config.Flute.ServerPort))+
		"/graphql?operationName=_&query="+queryURLEncoder(query), nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		// Check response
		respBody, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			result := string(respBody)
			if strings.Contains(result, "errors") {
				return "", errors.New(result)
			}
			return result, nil
		}

		return "", err
	}

	return "", errors.New("http response returned error code")
}

// ToFluteGetNodes : Get not activated nodes info from flute module
func ToFluteGetNodes() (ListNodeData, error) {
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

	client := &http.Client{Timeout: time.Duration(config.Flute.RequestTimeoutMs) * time.Millisecond}
	req, err := http.NewRequest("GET", "http://"+config.Flute.ServerAddress+":"+strconv.Itoa(int(config.Flute.ServerPort))+
		"/graphql?query="+queryURLEncoder(query), nil)
	if err != nil {
		return listNodeData, err
	}
	resp, err := client.Do(req)

	if err != nil {
		return listNodeData, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		// Check response
		respBody, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			str := string(respBody)

			if strings.Contains(str, "errors") {
				return listNodeData, errors.New(str)
			}

			err = json.Unmarshal([]byte(str), &listNodeData)
			if err != nil {
				return listNodeData, err
			}

			return listNodeData, nil
		}

		return listNodeData, err
	}

	return listNodeData, errors.New("http response returned error code")
}

// ToFluteUpdateNode : Add server_uuid information to each nodes
func ToFluteUpdateNode(node model.Node, serverUUID string) error {
	query := "mutation{\n" +
		"	update_node(uuid:\"" + node.UUID + "\", server_uuid:\"" + serverUUID + "\", active: 1){\n" +
		"		uuid\n" +
		"	}\n" +
		"}"

	client := &http.Client{Timeout: time.Duration(config.Flute.RequestTimeoutMs) * time.Millisecond}
	req, err := http.NewRequest("GET", "http://"+config.Flute.ServerAddress+":"+strconv.Itoa(int(config.Flute.ServerPort))+
		"/graphql?query="+queryURLEncoder(query), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		// Check response
		respBody, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			str := string(respBody)

			if strings.Contains(str, "errors") {
				return errors.New(str)
			}

			return nil
		}

		return err
	}

	return errors.New("http response returned error code")
}

// Cello

// ToCelloCreateDisk : Create os or data disk
func ToCelloCreateDisk(volume model.Volume, serverUUID string) error {
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

	client := &http.Client{Timeout: time.Duration(config.Cello.RequestTimeoutMs) * time.Millisecond}
	req, err := http.NewRequest("GET", "http://"+config.Cello.ServerAddress+":"+strconv.Itoa(int(config.Cello.ServerPort))+
		"/graphql?operationName=_&query="+queryURLEncoder(query), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		// Check response
		respBody, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			str := string(respBody)

			if strings.Contains(str, "errors") {
				return errors.New(str)
			}
			return nil
		}

		return err
	}

	return errors.New("http response returned error code")
}

// Harp

// ToHarpGetSubnet : Get subnet info from harp module
func ToHarpGetSubnet(subnetUUID string) (SubnetData, error) {
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

	client := &http.Client{Timeout: time.Duration(config.Harp.RequestTimeoutMs) * time.Millisecond}
	req, err := http.NewRequest("GET", "http://"+config.Harp.ServerAddress+":"+strconv.Itoa(int(config.Harp.ServerPort))+
		"/graphql?query="+queryURLEncoder(query), nil)
	if err != nil {
		return subnetData, err
	}
	resp, err := client.Do(req)

	if err != nil {
		return subnetData, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		// Check response
		respBody, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			str := string(respBody)

			if strings.Contains(str, "errors") {
				return subnetData, errors.New(str)
			}

			err = json.Unmarshal([]byte(str), &subnetData)
			if err != nil {
				return subnetData, err
			}

			return subnetData, nil
		}

		return subnetData, err
	}

	return subnetData, errors.New("http response returned error code")
}

// ToHarpUpdateSubnet : Add server_uuid to subnet
func ToHarpUpdateSubnet(subnetUUID string, serverUUID string) (SubnetData, error) {
	query := "mutation _ {\n" +
		"	update_subnet(uuid: \"" + subnetUUID + "\", server_uuid: \"" + serverUUID + "\"){\n" +
		"		uuid\n" +
		"		server_uuid\n" +
		"	}\n" +
		"}"

	var subnetData SubnetData

	client := &http.Client{Timeout: time.Duration(config.Harp.RequestTimeoutMs) * time.Millisecond}
	req, err := http.NewRequest("GET", "http://"+config.Harp.ServerAddress+":"+strconv.Itoa(int(config.Harp.ServerPort))+
		"/graphql?query="+queryURLEncoder(query), nil)
	if err != nil {
		return subnetData, err
	}
	resp, err := client.Do(req)

	if err != nil {
		return subnetData, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		// Check response
		respBody, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			str := string(respBody)

			if strings.Contains(str, "errors") {
				return subnetData, errors.New(str)
			}

			err = json.Unmarshal([]byte(str), &subnetData)
			if err != nil {
				return subnetData, err
			}

			return subnetData, nil
		}

		return subnetData, err
	}

	return subnetData, errors.New("http response returned error code")
}

// ToHarpCreateDHCPDConfig : Add server_uuid to subnet
func ToHarpCreateDHCPDConfig(subnetUUID string, nodeUUIDsStr string) error {
	query := "mutation _ {\n" +
		"	create_dhcpd_conf(subnet_uuid: \"" + subnetUUID + "\", node_uuids: \"" + nodeUUIDsStr + "\")\n" +
		"}"

	client := &http.Client{Timeout: time.Duration(config.Harp.RequestTimeoutMs) * time.Millisecond}
	req, err := http.NewRequest("GET", "http://"+config.Harp.ServerAddress+":"+strconv.Itoa(int(config.Harp.ServerPort))+"/graphql?query="+queryURLEncoder(query), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		// Check response
		respBody, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			str := string(respBody)

			if strings.Contains(str, "errors") {
				return errors.New(str)
			}

			logger.Logger.Println("ToHarpCreateDHCPDConfig: Successfully created dhcpd config for subnetUUID=" + subnetUUID)

			return nil
		}

		return err
	}

	return nil
}
