package graphql

import "time"

type ListNodeData struct {
	Data struct {
		ListNode []struct {
			Active      int       `json:"active"`
			BmcIP       string    `json:"bmc_ip"`
			BmcMacAddr  string    `json:"bmc_mac_addr"`
			CPUCores    int       `json:"cpu_cores"`
			CreatedAt   time.Time `json:"created_at"`
			Description string    `json:"description"`
			Memory      int       `json:"memory"`
			PxeMacAddr  string    `json:"pxe_mac_addr"`
			Status      string    `json:"status"`
			UUID        string    `json:"uuid"`
		} `json:"list_node"`
	} `json:"data"`
}