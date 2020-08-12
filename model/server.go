package model

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
