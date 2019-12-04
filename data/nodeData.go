package data

import "hcc/violin/model"

// AllNodeData : Data structure of all_node
type AllNodeData struct {
	Data struct {
		AllNode []model.Node `json:"all_node"`
	} `json:"data"`
}
