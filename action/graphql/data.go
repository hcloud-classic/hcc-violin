package graphql

import (
	"hcc/violin/model"
)

// UpdateNodeData : Data structure of update_node
type UpdateNodeData struct {
	Data struct {
		Node model.Node `json:"update_node"`
	} `json:"data"`
}

// ListNodeData : Data structure of list_node
type ListNodeData struct {
	Data struct {
		ListNode []model.Node `json:"list_node"`
	} `json:"data"`
}
