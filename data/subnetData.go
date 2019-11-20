package data

import "hcc/violin/model"

// SubnetData : Data structure of subnet
type SubnetData struct {
	Data struct {
		Subnet model.Subnet `json:"subnet"`
	} `json:"data"`
}
