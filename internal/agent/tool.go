package agent

type Tool struct {
	Name        string
	Description string
	Parameters  map[string]any
}

type ToolError struct {
	Error string `json:"error"`
}
