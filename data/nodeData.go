package data

import "hcc/violin/model"

// AllNodeData : Data structure of all_node
type AllNodeData struct {
	Data struct {
		AllNode []model.Node `json:"all_node"`
	} `json:"data"`
}

type SingleNodeData struct {
	Data struct {
		SingleNode model.Node `json:"node"`
	} `json:"data"`
}

type ScheduledNodeData struct {
	Data struct {
		ScheduledNode model.ScheduledNodes `json:"schedule_nodes"`
	} `json:"data"`
}

// type ScheduledNodeData struct {
// 	Data struct {
// 		NodeList []string `json:"selected_nodes"`
// 	} `json:"data"`
// }
