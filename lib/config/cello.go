package config

type cello struct {
	ServerAddress    string `goconf:"cello:cello_server_address"`     // ServerAddress : IP address of server which installed cello module
	ServerPort       int64  `goconf:"cello:cello_server_port"`        // ServerPort : Listening port number of cello module
	RequestTimeoutMs int64  `goconf:"cello:cello_request_timeout_ms"` // RequestTimeoutMs : HTTP timeout for GraphQL request to cello module
}

// Cello : cello config structure
var Cello cello
