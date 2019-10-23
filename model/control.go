package model

// Control : Struct of Control
type Control struct {
	HccIPRange string `json:"iprange"`
	HccCommand string `json:"action"`
}

// Controls : Array struct of Control
type Controls struct {
	Controls Control `json:"control"`
}
