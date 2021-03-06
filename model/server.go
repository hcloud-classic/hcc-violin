package model

import "time"

// Server - cgs
type Server struct {
	UUID       string    `json:"uuid"`
	SubnetUUID string    `json:"subnet_uuid"`
	OS         string    `json:"os"`
	ServerName string    `json:"server_name"`
	ServerDesc string    `json:"server_desc"`
	CPU        int       `json:"cpu"`
	Memory     int       `json:"memory"`
	DiskSize   int       `json:"disk_size"`
	Status     string    `json:"status"`
	UserUUID   string    `json:"user_uuid"`
	CreatedAt  time.Time `json:"created_at"`
}

// Servers - cgs
type Servers struct {
	Server []Server `json:"server"`
}

// ServerNum - cgs
type ServerNum struct {
	Number int `json:"number"`
}

//Vnc : For Vnc Information
type Vnc struct {
	ServerUUID     string `json:"server_uuid"`
	TargetIP       string `json:"target_ip"`
	TargetPort     string `json:"target_port"`
	WebSocket      string `json:"websocket_port"`
	TargetPass     string `json:"target_pass"`
	Info           string `json:"vnc_info"`
	ActionClassify string `json:"action"`
}
