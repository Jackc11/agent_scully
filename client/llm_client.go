package client

import (
	"agent_scully/tools"
	"agent_scully/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	BackendOllama = 0
	BackendVLLM   = 1
)

type Message struct {
	Role       string           `json:"role"`
	Content    string           `json:"content"`
	ToolCall   []tools.ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type LLMClient struct {
	history []Message
	tools   []tools.ToolDefinition
	conf    *utils.AgentConfig
	client  *http.Client
}

func CreateLLMClient(conf *utils.AgentConfig) *LLMClient {
	return &LLMClient{
		conf:  conf,
		tools: tools.CreateToolDefition(),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func createOllamaRequest(c *utils.AgentConfig, h []Message, t []tools.ToolDefinition) (string, []byte, error) {
	req := struct {
		Model    string                 `json:"model"`
		Messages []Message              `json:"messages"`
		Stream   bool                   `json:"stream"`
		Tools    []tools.ToolDefinition `json:"tools,omitempty"`
		Options  utils.OllamaOptions    `json:"options,omitempty"`
	}{
		Model:    c.OllamaConfig.Model,
		Messages: h,
		Stream:   c.OllamaConfig.Stream,
		Tools:    t,
		Options:  c.OllamaConfig.Options,
	}

	body, err := json.Marshal(req)
	return c.OllamaConfig.Url, body, err
}

func createVllmRequest(c *utils.AgentConfig, h []Message, t []tools.ToolDefinition) (string, []byte, error) {
	req := struct {
		Model    string                 `json:"model"`
		Messages []Message              `json:"messages"`
		Stream   bool                   `json:"stream"`
		Tools    []tools.ToolDefinition `json:"tools,omitempty"`
	}{
		Model:    c.VllmConfig.Model,
		Messages: h,
		Stream:   false,
		Tools:    t,
	}

	body, err := json.Marshal(req)
	return c.VllmConfig.Url, body, err
}

func buildRequestPayload(conf *utils.AgentConfig, history []Message, tools []tools.ToolDefinition) (string, []byte, error) {
	switch conf.LLMBackend {
	case BackendOllama:
		return createOllamaRequest(conf, history, tools)
	default:
		return createVllmRequest(conf, history, tools)
	}
}

func getLLMResponse(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API error (%d): %s", resp.StatusCode, string(body))
	}

	return body, err
}

func parseResponseMessage(conf *utils.AgentConfig, body []byte) (Message, error) {
	if conf.LLMBackend == BackendVLLM {
		var vllmResp struct {
			Choices []struct {
				Message Message `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(body, &vllmResp); err != nil {
			return Message{}, err
		}
		if len(vllmResp.Choices) == 0 {
			return Message{}, fmt.Errorf("vLLM returned empty choices")
		}
		return vllmResp.Choices[0].Message, nil
	}

	// Default: Ollama response format
	var ollamaResp struct {
		Message Message `json:"message"`
	}
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return Message{}, err
	}
	return ollamaResp.Message, nil
}

func (c *LLMClient) RunPrompt(prompt string) (string, error) {
	c.history = append(c.history, Message{
		Role:    "user",
		Content: prompt,
	})

	var usedToolOutputs []string

	for range 6 {
		url, payload, err := buildRequestPayload(c.conf, c.history, c.tools)
		if err != nil {
			return "", fmt.Errorf("failed to create request payload: %v", err)
		}
		if url == "" {
			return "", fmt.Errorf("LLM endpoint URL is empty for backend %d", c.conf.LLMBackend)
		}

		resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			return "", fmt.Errorf("HTTP request failed: %v", err)
		}

		body, err := getLLMResponse(resp)
		if err != nil {
			return "", err
		}

		assistantMsg, err := parseResponseMessage(c.conf, body)
		if err != nil {
			return "", fmt.Errorf("failed to parse JSON: %v", err)
		}

		c.history = append(c.history, assistantMsg)

		if len(assistantMsg.ToolCall) == 0 {
			return assistantMsg.Content, nil
			// return assistantMsg.Content + extractToolSources(usedToolOutputs), nil
		}

		for _, tc := range assistantMsg.ToolCall {
			args := tc.Function.ParsedArguments()
			var toolOutput string

			switch tc.Function.Name {
			case "web_search":
				query, _ := args["query"].(string)
				out, err := tools.SearchWeb(query)
				if err != nil {
					toolOutput = fmt.Sprintf("Search error: %v", err)
				} else {
					toolOutput = out
					usedToolOutputs = append(usedToolOutputs, out)
				}

			case "read_webpage":
				targetURL, _ := args["url"].(string)
				out, err := tools.ReadWebpage(targetURL)
				if err != nil {
					toolOutput = fmt.Sprintf("Read webpage error: %v", err)
				} else {
					toolOutput = out
					usedToolOutputs = append(usedToolOutputs, out)
				}

			case "curl":
				target, _ := args["target"].(string)
				target = strings.TrimSpace(target)
				if target == "" {
					toolOutput = "Curl error: target is required"
				} else {
					toolOutput = fmt.Sprintf("curl %s output:\n%s", target, tools.CurlCommand(target))
				}

			case "ping":
				target, _ := args["target"].(string)
				target = strings.TrimSpace(target)
				if target == "" {
					toolOutput = "Ping error: target is required"
				} else {
					toolOutput = fmt.Sprintf("ping %s output:\n%s", target, tools.PingCommand(target))
				}

			case "nslookup":
				target, _ := args["target"].(string)
				target = strings.TrimSpace(target)
				if target == "" {
					toolOutput = "Nslookup error: target is required"
				} else {
					toolOutput = fmt.Sprintf("nslookup %s output:\n%s", target, tools.NslookupCommand(target))
				}

			default:
				toolOutput = fmt.Sprintf("Unknown tool: %s", tc.Function.Name)
			}

			c.history = append(c.history, Message{
				Role:       "tool",
				Content:    toolOutput,
				ToolCallID: tc.ID,
			})
		}
	}

	return "", fmt.Errorf("agent reached maximum tool iterations without producing a final answer")
}

func extractToolSources(toolOutputs []string) string {
	var sources []string
	for _, out := range toolOutputs {
		lines := strings.Split(out, "\n")
		for i := 0; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if strings.HasPrefix(line, "[") && i+1 < len(lines) && strings.HasPrefix(lines[i+1], "URL: ") {
				rawURL := strings.TrimPrefix(lines[i+1], "URL: ")
				sources = append(sources, fmt.Sprintf("- %s — <%s>", line, rawURL))
			} else if strings.HasPrefix(line, "Source: http") {
				rawURL := strings.TrimPrefix(line, "Source: ")
				sources = append(sources, fmt.Sprintf("- <%s>", rawURL))
			}
		}
	}
	if len(sources) == 0 {
		return ""
	}
	return "\n\n### Sources\n" + strings.Join(sources, "\n")
}
