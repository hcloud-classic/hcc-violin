package model

// NormalAction :
type NormalAction struct {
}

// HccAction : specified Hcloud commands
type HccAction struct {
	ActionArea  string `json:"action_area"`  // Nodes for Cluster realem
	ActionClass string `json:"action_class"` //add, del, status, poweroff, reboot
	ActionScope string `json:"action_scope"` // n, n:n+2, if n is '0' all node add
	HccIPRange  string `json:"iprange"`      // range xxx.xxx.xxx.xxx yyy.yyy.yyy.yyy
	ServerUUID  string `json:"server_uuid"`
}

//Action : Any Action
type Action struct {
	//ActionType is Classified type of action that Executable shell command name ex) ls or cp or mkdir...
	// hcc ,normal(Precompleted Action)
	// Ex )
	// hcc =>  cludter control action
	// normal => ls -al
	// normal => scp xxx root@123.456.789.123:/root/
	ActionType   string       `json:"action_type"`
	NormalType   NormalAction `json:"normal_action"`
	HccType      HccAction    `json:"hcc_action"`
	ActionResult string       `json:"action_result"`
}

// Control : Struct of Control
type Control struct {
	Control   Action `json:"action"`
	Publisher string `json:"publisher"` //Who send action.
	Receiver  string `json:"receiver"`  //Who has receive the result of action.
}

// Controls : Array struct of Control
type Controls struct {
	Controls Control `json:"control"`
}
