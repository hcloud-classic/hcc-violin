package graphql

import (
	"encoding/json"
	"errors"
	"fmt"
	"hcc/violin/lib/config"
	"io/ioutil"
	"net/http"
	"time"
)

// GetNodes : Get nodes info from flute module
func GetNodes() (ListNodeData, error) {
	var listNodeData ListNodeData

	client := &http.Client{Timeout: time.Duration(config.HTTP.RequestTimeoutMs) * time.Millisecond}
	req, err := http.NewRequest("GET", "http://192.168.110.99:7000/graphql?query=query%20%7B%0A%20%20list_node(active%3A%200%2C%20row%3A10%2C%20page%3A1)%20%7B%0A%20%20%20%20uuid%0A%20%20%20%20bmc_mac_addr%0A%20%20%20%20bmc_ip%0A%20%20%20%20pxe_mac_addr%0A%20%20%20%20status%0A%20%20%20%20cpu_cores%0A%20%20%20%20memory%0A%20%20%20%20description%0A%20%20%20%20created_at%0A%20%20%20%20active%0A%20%20%7D%0A%7D%0A", nil)
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
