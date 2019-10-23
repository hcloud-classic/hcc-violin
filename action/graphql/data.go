package graphql

import (
	"hcc/violin/model"
)

// Flute

// ListNodeData : Data structure of list_node
type ListNodeData struct {
	Data struct {
		ListNode []model.Node `json:"list_node"`
	} `json:"data"`
}
