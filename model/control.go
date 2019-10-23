package model

// Control : Struct of Control
type Control struct {
	HccCommand string `json:"action"`
	HccIPRange string `json:"iprange"`
}

// Controls : Array struct of Control
type Controls struct {
	Controls Control `json:"control"`
}
