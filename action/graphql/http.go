package graphql

import (
	"encoding/json"
	"errors"
	"fmt"
	"hcc/violin/lib/config"
	"hcc/violin/model"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// Flute

// GetNodes : Get not activated nodes info from flute module
func GetNodes() (ListNodeData, error) {
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
			fmt.Println(str)
			if err != nil {
				return listNodeData, err
			}

			return listNodeData, nil
		}

		return listNodeData, err
	}

	return listNodeData, errors.New("http response returned error code")
}

func UpdateNode(node model.Node, serverUUID string) error {
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

func CreateDisk(volume model.Volume, serverUUID string) error {
	client := &http.Client{Timeout: time.Duration(config.Cello.RequestTimeoutMs) * time.Millisecond}
	req, err := http.NewRequest("GET", "http://"+config.Cello.ServerAddress+":"+strconv.Itoa(int(config.Cello.ServerPort))+
	"graphql?operationName=_&query=mutation%20_%20%7B%0A%20%20create_volume(size%3A" + strconv.Itoa(volume.Size) + "%2C%20filesystem%3A%22" + volume.Filesystem + "%22%2C%20server_uuid%3A%22"+ serverUUID +"%22%2C%20use_type%3A%22" +volume.UseType +"%22%2C%20user_uuid%3A%22" +volume.UserUUID + "%22)%20%7B%0A%20%20%20%20uuid%0A%20%20%20%20size%0A%20%20%20%20filesystem%0A%20%20%20%20server_uuid%0A%20%20%20%20use_type%0A%20%20%20%20user_uuid%0A%20%20%20%20created_at%0A%20%20%7D%0A%7D", nil)
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
