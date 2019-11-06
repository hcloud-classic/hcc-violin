package config

type http struct {
	Port             int64 `goconf:"http:port"`               // Port : Port number for listening graphql request via http server
	RequestTimeoutMs int64 `goconf:"http:request_timeout_ms"` // RequestTimeoutMs : Timeout for HTTP request
}

// HTTP : http config structure
var HTTP http
