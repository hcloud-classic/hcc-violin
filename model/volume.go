package model

import "time"

// OSDiskSize : Disk size for OS use
var OSDiskSize = 100

// Volume - cgs
type Volume struct {
	UUID       string    `json:"uuid"`
	Size       int       `json:"size"`
	Filesystem string    `json:"filesystem"`
	ServerUUID string    `json:"server_uuid"`
	UseType    string    `json:"use_type"`
	UserUUID   string    `json:"user_uuid"`
	CreatedAt  time.Time `json:"created_at"`
	NetworkIP  string    `json:"network_ip"`
	GatewayIP  string    `json:"gateway_ip"`
}

// Volumes - cgs
type Volumes struct {
	Volumes []Volume `json:"volume"`
}

// VolumeNum - cgs
type VolumeNum struct {
	Number int `json:"number"`
}
