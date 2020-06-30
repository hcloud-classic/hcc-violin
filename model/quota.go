package model

type Quota struct {
	ServerUUID    string
	CPU           int
	Memory        int
	NumberOfNodes int
}
