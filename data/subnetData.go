package data

import "hcc/violin/model"

type SubnetData struct {
	Data struct {
		Subnet model.Subnet `json:"subnet"`
	} `json:"data"`
}
