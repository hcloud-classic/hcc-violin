package config

type autoscale struct {
	Debug                              string `goconf:"autoscale:debug"`                                  // Debug : Enable debug logs for AutoScale
	CheckServerResourceIntervalMs      int64  `goconf:"autoscale:check_server_resource_interval_ms"`      // CheckServerResourceIntervalMs : Server resource check interval (ms)
	AutoScaleTriggerCPUUsagePercent    int64  `goconf:"autoscale:autoscale_trigger_cpu_usage_percent"`    // AutoScaleTriggerCPUUsagePercent : CPU usage percent of triggering auto-scale
	AutoScaleTriggerMemoryUsagePercent int64  `goconf:"autoscale:autoscale_trigger_memory_usage_percent"` // AutoScaleTriggerMemoryUsagePercent : Memory usage percent of triggering auto-scale
}

// AutoScale : autoscale config structure
var AutoScale autoscale
