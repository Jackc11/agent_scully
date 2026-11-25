package tools

import "encoding/json"

type FunctionCall struct {
	Name string `json:"name"`
	// Arguments map[string]interface{} `json:"arguments"`
	Arguments json.RawMessage `json:"arguments"`
}

type ToolCall struct {
	ID       string       `json:"id,omitempty"`   // Required by vllm
	Type     string       `json:"type,omitempty"` // required by vllm
	Function FunctionCall `json:"function"`
}

type ToolDefinition struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

type FunctionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

func (fc FunctionCall) ParsedArguments() map[string]interface{} {
	var args map[string]interface{}
	if len(fc.Arguments) == 0 {
		return args
	}
	// If vLLM returned a quoted JSON string ("{\"query\":\"foo\"}"), unquote it first
	if fc.Arguments[0] == '"' {
		var unquoted string
		if err := json.Unmarshal(fc.Arguments, &unquoted); err == nil {
			_ = json.Unmarshal([]byte(unquoted), &args)
			return args
		}
	}
	// Otherwise unmarshal directly (Ollama object format)
	_ = json.Unmarshal(fc.Arguments, &args)
	return args
}

func CreateToolDefition() []ToolDefinition {
	return []ToolDefinition{
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "web_search",
				Description: "Search the web using DuckDuckGo for current events, facts, or URLs. Returns numbered sources with URLs that you must cite in your answer.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "The search query string.",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "read_webpage",
				Description: "Open a specific webpage URL and read the full text content inside it. Cite the returned URL when you use its information.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"url": map[string]any{
							"type":        "string",
							"description": "The full http or https URL to read.",
						},
					},
					"required": []string{"url"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "curl",
				Description: "Run curl with verbose headers and a 10 second timeout against a URL or host to inspect HTTP response headers and connection details.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"target": map[string]any{
							"type":        "string",
							"description": "The URL, domain, or IP address to inspect with curl.",
						},
					},
					"required": []string{"target"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "ping",
				Description: "Ping a hostname or IP address 5 times to test network reachability and latency.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"target": map[string]any{
							"type":        "string",
							"description": "The hostname or IP address to ping.",
						},
					},
					"required": []string{"target"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "nslookup",
				Description: "Query DNS records for a hostname or IP address.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"target": map[string]any{
							"type":        "string",
							"description": "The hostname or IP address to look up.",
						},
					},
					"required": []string{"target"},
				},
			},
		},
	}
}
