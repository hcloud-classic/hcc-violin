package model

import "time"

type Subnet struct {
	UUID           string    `json:"uuid"`
	NetworkIP      string    `json:"network_ip"`
	Netmask        string    `json:"netmask"`
	Gateway        string    `json:"gateway"`
	NextServer     string    `json:"next_server"`
	NameServer     string    `json:"name_server"`
	DomainName     string    `json:"domain_name"`
	ServerUUID     string    `json:"server_uuid"`
	LeaderNodeUUID string    `json:"leader_node_uuid"`
	OS             string    `json:"os"`
	SubnetName     string    `json:"subnet_name"`
	CreatedAt      time.Time `json:"created_at"`
}

type Subnets struct {
	Subnets []Subnet `json:"subnet"`
}

type SubnetNum struct {
	Number int `json:"number"`
}
