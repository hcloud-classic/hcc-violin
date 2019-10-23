package graphql

import (
	"hcc/violin/model"
)

type ListNodeData struct {
	Data struct {
		ListNode[] model.Node `json:"list_node"`
	} `json:"data"`
}