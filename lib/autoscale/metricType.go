package autoscale

type metric []struct {
	Series []struct {
		Name    string            `json:"name,omitempty"`
		Tags    map[string]string `json:"tags,omitempty"`
		Columns []string          `json:"columns,omitempty"`
		Values  [][]interface{}   `json:"values,omitempty"`
		Partial bool              `json:"partial,omitempty"`
	} `json:"Series"`
	Messages []*struct {
		Level string
		Text  string
	} `json:"Messages"`
	Err string `json:"error,omitempty"`
}
