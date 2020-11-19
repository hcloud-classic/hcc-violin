package http

import (
	"encoding/json"
	"errors"
	"fmt"
	violinData "hcc/violin/data"
	"hcc/violin/lib/config"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func getModuleHTTPInfo(moduleName string) (time.Duration, string, error) {
	var timeout time.Duration
	var url = "http://"
	switch moduleName {
	case "flute":
		timeout = time.Duration(config.Flute.RequestTimeoutMs)
		url += config.Flute.ServerAddress + ":" + strconv.Itoa(int(config.Flute.ServerPort))
		break
	case "harp":
		timeout = time.Duration(config.Harp.RequestTimeoutMs)
		url += config.Harp.ServerAddress + ":" + strconv.Itoa(int(config.Harp.ServerPort))
		break
	case "cello":
		timeout = time.Duration(config.Cello.RequestTimeoutMs)
		url += config.Cello.ServerAddress + ":" + strconv.Itoa(int(config.Cello.ServerPort))
		break
	case "violin_scheduler":
		timeout = time.Duration(config.ViolinScheduler.RequestTimeoutMs)
		url += config.ViolinScheduler.ServerAddress + ":" + strconv.Itoa(int(config.ViolinScheduler.ServerPort))
		break
	default:
		return 0, "", errors.New("unknown module name")
	}

	return timeout, url, nil
}

// DoHTTPRequest : Send http request to other modules with GraphQL query string.
func DoHTTPRequest(moduleName string, needData bool, data interface{}, query string, singleData bool) (interface{}, error) {
	timeout, url, err := getModuleHTTPInfo(moduleName)
	if err != nil {
		return nil, err
	}

	url += "/graphql?query=" + queryURLEncoder(query)
	client := &http.Client{Timeout: timeout * time.Millisecond}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			result := string(respBody)

			if strings.Contains(result, "errors") {
				return nil, errors.New(result)
			}
			if needData {
				if data == nil {
					return nil, errors.New("needData marked as true but data is nil")
				}

				switch moduleName {
				case "flute":

					if singleData {
						NodeData := data.(violinData.SingleNodeData)
						err = json.Unmarshal([]byte(result), &NodeData)

						if err != nil {
							return nil, err
						}
						fmt.Println("Flute SingleNodeData", NodeData)
						return NodeData, nil
					} else {
						NodeData := data.(violinData.AllNodeData)
						err = json.Unmarshal([]byte(result), &NodeData)
						if err != nil {
							return nil, err
						}
						fmt.Println("Flute allNodeData", NodeData)
						return NodeData, nil
					}

				case "harp":
					subnetData := data.(violinData.SubnetData)
					err = json.Unmarshal([]byte(result), &subnetData)
					if err != nil {
						return nil, err
					}

					return subnetData, nil
				case "violin_scheduler":

					scheduledList := data.(violinData.ScheduledNodeData)
					err = json.Unmarshal([]byte(result), &scheduledList)
					if err != nil {
						return nil, err
					}

					return scheduledList, nil
				case "violin_novnc":
					vncData := data.(violinData.VncNodeData)
					fmt.Println("result : ", result)
					err = json.Unmarshal([]byte(result), &vncData)
					if err != nil {
						return nil, err
					}
					fmt.Println("vncData : ", vncData)

					return vncData, nil

				default:
					return nil, errors.New("data is not supported for " + moduleName + " module")
				}
			}

			return result, nil
		}

		return nil, err
	}

	return nil, errors.New("http response returned error code")
}
