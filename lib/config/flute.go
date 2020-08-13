package config

type flute struct {
	ServerAddress               string `goconf:"flute:flute_server_address"`                   // ServerAddress : IP address of server which installed flute module
	ServerPort                  int64  `goconf:"flute:flute_server_port"`                      // ServerPort : Listening port number of flute module
	RequestTimeoutMs            int64  `goconf:"flute:flute_request_timeout_ms"`               // RequestTimeoutMs : HTTP timeout for GraphQL request to flute module
	WaitForLeaderNodeTimeoutSec int64  `goconf:"flute:flute_wait_for_leader_node_timeout_sec"` // WaitForLeaderNodeTimeoutSec : Waiting timeout for turn on compute nodes after leader node turned on
}

// Flute : flute config structure
var Flute flute
