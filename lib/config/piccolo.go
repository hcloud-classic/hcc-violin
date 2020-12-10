package config

type piccolo struct {
	ServerAddress    string `goconf:"piccolo:piccolo_server_address"`     // ServerAddress : IP address of server which installed piccolo module
	ServerPort       int64  `goconf:"piccolo:piccolo_server_port"`        // ServerPort : Listening port number of piccolo module
	RequestTimeoutMs int64  `goconf:"piccolo:piccolo_request_timeout_ms"` // RequestTimeoutMs : Timeout for gRPC request to piccolo module
}

// Piccolo : Piccolo config structure
var Piccolo piccolo
