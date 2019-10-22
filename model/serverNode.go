package model

import "time"

// ServerNode :
type ServerNode struct {
	UUID       string    `json:"uuid"`
	ServerUUID string    `json:"server_uuid"`
	NodeUUID   string    `json:"node_uuid"`
	CreatedAt  time.Time `json:"created_at"`
}

// ServerNodes :
type ServerNodes struct {
	Server []Server `json:"server_node"`
}
