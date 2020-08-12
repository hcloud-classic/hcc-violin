package data

import "hcc/violin/model"

// ScheduledNodeData :
type ScheduledNodeData struct {
	Data struct {
		ScheduledNode model.ScheduledNodes `json:"schedule_nodes"`
	} `json:"data"`
}

// VncNodeData :
type VncNodeData struct {
	Data struct {
		ScheduledNode model.Vnc `json:"control_vnc"`
	} `json:"data"`
}

// type ScheduledNodeData struct {
// 	Data struct {
// 		NodeList []string `json:"selected_nodes"`
// 	} `json:"data"`
// }
