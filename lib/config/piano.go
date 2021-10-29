package config

type piano struct {
	ServerAddress    string `goconf:"piano:pianon_server_address"`    // ServerAddress : IP address of server which installed piano module
	ServerPort       int64  `goconf:"piano:piano_server_port"`        // ServerPort : Listening port number of piano module
	RequestTimeoutMs int64  `goconf:"piano:piano_request_timeout_ms"` // RequestTimeoutMs : HTTP timeout for gRPC request to piano module
}

// Piano : piano config structure
var Piano piano
