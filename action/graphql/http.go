package graphql

import (
	"encoding/json"
	"errors"
	"fmt"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"hcc/violin/model"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// Flute

// ToFluteOnNode : Turn on the node by sending WOL magic packet
func ToFluteOnNode(macAddr string) (string, error) {
	client := &http.Client{Timeout: time.Duration(config.Flute.RequestTimeoutMs) * time.Millisecond}
	req, err := http.NewRequest("GET", "http://"+config.Flute.ServerAddress+":"+strconv.Itoa(int(config.Flute.ServerPort))+
		"/graphql?operationName=_&query=mutation%20_%20%7B%0A%20%20on_node(mac%3A%20%22"+macAddr+"%22)%0A%7D&variables=%7B%7D", nil)
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
			return result, nil
		}

		return "", err
	}

	return "", errors.New("http response returned error code")
}

// ToFluteGetNodes : Get not activated nodes info from flute module
func ToFluteGetNodes() (ListNodeData, error) {
	var listNodeData ListNodeData

	client := &http.Client{Timeout: time.Duration(config.Flute.RequestTimeoutMs) * time.Millisecond}
	req, err := http.NewRequest("GET", "http://"+config.Flute.ServerAddress+":"+strconv.Itoa(int(config.Flute.ServerPort))+
		"/graphql?query=query%20%7B%0A%20%20list_node(active%3A%200%2C%20row%3A10%2C%20page%3A1)%20%7B%0A%20%20%20%20uuid%0A%20%20%20%20bmc_mac_addr%0A%20%20%20%20bmc_ip%0A%20%20%20%20pxe_mac_addr%0A%20%20%20%20status%0A%20%20%20%20cpu_cores%0A%20%20%20%20memory%0A%20%20%20%20description%0A%20%20%20%20created_at%0A%20%20%20%20active%0A%20%20%7D%0A%7D%0A", nil)
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
	client := &http.Client{Timeout: time.Duration(config.Flute.RequestTimeoutMs) * time.Millisecond}
	req, err := http.NewRequest("GET", "http://"+config.Flute.ServerAddress+":"+strconv.Itoa(int(config.Flute.ServerPort))+
		"/graphql?query=mutation%7B%0A%20%20update_node(uuid%3A%22"+node.UUID+"%22%2C%20server_uuid%3A%22"+serverUUID+"%22%2C%20active%3A%20"+strconv.Itoa(1)+")%7B%0A%20%20%20%20uuid%0A%20%20%7D%0A%7D", nil)
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
			fmt.Println(str)

			return nil
		}

		return err
	}

	return errors.New("http response returned error code")
}

// Cello

// ToCelloCreateDisk : Create os or data disk
func ToCelloCreateDisk(volume model.Volume, serverUUID string) error {
	client := &http.Client{Timeout: time.Duration(config.Cello.RequestTimeoutMs) * time.Millisecond}
	req, err := http.NewRequest("GET", "http://"+config.Cello.ServerAddress+":"+strconv.Itoa(int(config.Cello.ServerPort))+
		"/graphql?operationName=_&query=mutation%20_%20%7B%0A%20%20create_volume(size%3A"+strconv.Itoa(volume.Size)+"%2C%20filesystem%3A%22"+volume.Filesystem+"%22%2C%20server_uuid%3A%22"+serverUUID+"%22%2C%20use_type%3A%22"+volume.UseType+"%22%2C%20user_uuid%3A%22"+volume.UserUUID+"%22)%20%7B%0A%20%20%20%20uuid%0A%20%20%20%20size%0A%20%20%20%20filesystem%0A%20%20%20%20server_uuid%0A%20%20%20%20use_type%0A%20%20%20%20user_uuid%0A%20%20%20%20created_at%0A%20%20%7D%0A%7D", nil)
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
		_, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			return nil
		}

		return err
	}

	return errors.New("http response returned error code")
}

// Harp

// ToHarpGetSubnet : Get subnet info from harp module
func ToHarpGetSubnet(subnetUUID string) (SubnetData, error) {
	var subnetData SubnetData

	client := &http.Client{Timeout: time.Duration(config.Harp.RequestTimeoutMs) * time.Millisecond}
	req, err := http.NewRequest("GET", "http://"+config.Harp.ServerAddress+":"+strconv.Itoa(int(config.Harp.ServerPort))+
		"/graphql?query=query%20%7B%0A%20%20subnet(uuid%3A%22"+subnetUUID+"%22)%7B%0A%20%20%20%20uuid%0A%09network_ip%0A%20%20%20%20netmask%0A%20%20%20%20gateway%0A%20%20%20%20next_server%0A%20%20%20%20name_server%0A%20%20%20%20domain_name%0A%20%20%20%20server_uuid%0A%20%20%20%20leader_node_uuid%0A%20%20%20%20os%0A%20%20%20%20subnet_name%0A%20%20%20%20created_at%0A%20%20%7D%0A%7D", nil)
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
	var subnetData SubnetData

	client := &http.Client{Timeout: time.Duration(config.Harp.RequestTimeoutMs) * time.Millisecond}
	req, err := http.NewRequest("GET", "http://"+config.Harp.ServerAddress+":"+strconv.Itoa(int(config.Harp.ServerPort))+
		"/graphql?query=mutation%20_%20%7B%0A%20%20update_subnet(uuid%3A%20%22"+subnetUUID+"%22%2C%20server_uuid%3A%20%22"+serverUUID+"%22)%7B%0A%20%20%20%20uuid%0A%20%20%20%20server_uuid%0A%20%20%7D%0A%7D%0A&operationName=_", nil)
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
func ToHarpCreateDHCPDConfig(subnetUUID string, nodeUUIDsStr string) (error) {
	client := &http.Client{Timeout: time.Duration(config.Harp.RequestTimeoutMs) * time.Millisecond}
	req, err := http.NewRequest("GET", "http://" + config.Harp.ServerAddress + ":" + strconv.Itoa(int(config.Harp.ServerPort)) + "/graphql?query=mutation%20_%20%7B%0A%20%20create_dhcpd_conf(subnet_uuid%3A%20%22" + subnetUUID + "%22%2C%20node_uuids%3A%20%22" + nodeUUIDsStr + "%22)%0A%7D%0A&operationName=_", nil)
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
		_, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			logger.Logger.Println("ToHarpCreateDHCPDConfig: Successfully created dhcpd config for subnetUUID=" + subnetUUID)

			return  nil
		}

		return  err
	}

	return nil
}