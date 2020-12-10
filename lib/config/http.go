package config

type http struct {
	RequestTimeoutMs int64 `goconf:"http:request_timeout_ms"` // RequestTimeoutMs : Timeout for HTTP request
}

// HTTP : http config structure
var HTTP http
