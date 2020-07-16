package data

import "hcc/violin/model"

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

type VncNodeData struct {
	Data struct {
		ScheduledNode model.Vnc `json:"control_vnc"`
	} `json:"data"`
}
