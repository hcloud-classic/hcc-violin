package config

type violin_scheduler struct {
	ServerAddress    string `goconf:"violin_scheduler:violin_scheduler_server_address"`     // ServerAddress : IP address of server which installed harp module
	ServerPort       int64  `goconf:"violin_scheduler:violin_scheduler_server_port"`        // ServerPort : Listening port number of harp module
	RequestTimeoutMs int64  `goconf:"violin_scheduler:violin_scheduler_request_timeout_ms"` // RequestTimeoutMs : HTTP timeout for GraphQL request to harp module
}

// ViolinScheduler : ViolinScheduler
var ViolinScheduler violin_scheduler
