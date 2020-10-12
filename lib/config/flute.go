package config

type flute struct {
	ServerAddress               string `goconf:"flute:flute_server_address"`                   // ServerAddress : IP address of server which installed flute module
	ServerPort                  int64  `goconf:"flute:flute_server_port"`                      // ServerPort : Listening port number of flute module
	RequestTimeoutMs            int64  `goconf:"flute:flute_request_timeout_ms"`               // RequestTimeoutMs : HTTP timeout for GraphQL request to flute module
	TurnOffNodesWaitTimeSec     int64  `goconf:"flute:flute_turn_off_nodes_wait_time_sec"`     // TurnOffNodesWaitTimeSec : Wait time of turning of nodes
	TurnOffNodesRetryCounts     int64  `goconf:"flute:flute_turn_off_nodes_retry_counts"`      // TurnOffNodesRetryCounts : Retry counts of turning off nodes
	TurnOnNodesRetryCounts      int64  `goconf:"flute:flute_turn_on_nodes_retry_counts"`       // TurnOnNodesRetryCounts : Retry counts of turning on nodes
	WaitForLeaderNodeTimeoutSec int64  `goconf:"flute:flute_wait_for_leader_node_timeout_sec"` // WaitForLeaderNodeTimeoutSec : Waiting timeout for turn on compute nodes after leader node turned on
}

// Flute : flute config structure
var Flute flute
