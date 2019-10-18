package model

import "time"

// Subnet : Struct of Subnet
type Subnet struct {
	UUID           string    `json:"uuid"`
	NetworkIP      string    `json:"network_ip"`
	Netmask        string    `json:"netmask"`
	Gateway        string    `json:"gateway"`
	NextServer     string    `json:"next_server"`
	NameServer     string    `json:"name_server"`
	DomainName     string    `json:"domain_name"`
	MaxNodes       int       `json:"max_nodes"`
	NodeUUIDs      []string  `json:"node_uuids"`
	LeaderNodeUUID string    `json:"leader_node_uuid"`
	OS             string    `json:"os"`
	Name           string    `json:"name"`
	CreatedAt      time.Time `json:"created_at"`
}

// Subnets : Array struct of subnets
type Subnets struct {
	Subnets []Subnet `json:"subnet"`
}
