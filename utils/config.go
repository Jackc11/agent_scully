package utils

import (
	"encoding/json"
	"os"
)

const Role = "system"
const Content = `You are a helpful terminal assistant with access to web_search, read_webpage, curl, ping, and nslookup tools.

Tool Usage Rules:
- ONLY use web_search and read_webpage when the user explicitly tells you to search the web (e.g., "search the web", "search online", "look up online"). Otherwise, answer directly without calling web tools.
- Use curl, ping, and nslookup when asked to inspect HTTP headers, host reachability, or DNS records.

Citation & Sources Rules (STRICT):
- When you use web_search or read_webpage, you MUST include the exact full URLs from the tool output in your response.
- Cite facts inline using numbered markers like, [2].
- Always end your response with a "Sources:" section listing every source with its full raw URL (do NOT hide URLs inside markdown links).

Example format when web_search is used:
Go 1.24 was released with generic type aliases.

Sources:
[1] Go 1.24 Release Notes - https://go.dev`

type OllamaOptions struct {
	Temperature      *float64 `json:"temperature,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	TopK             *int     `json:"top_k,omitempty"`
	MinP             *float64 `json:"min_p,omitempty"`
	Seed             *int     `json:"seed,omitempty"`
	Stop             []string `json:"stop,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	NumCtx           *int     `json:"num_ctx,omitempty"`
	RepeatLastN      *int     `json:"repeat_last_n,omitempty"`
	Mirostat         *int     `json:"mirostat,omitempty"`
	NumGPU           *int     `json:"num_gpu,omitempty"`
	NumPredict       *int     `json:"num_predict,omitempty"`
	RepeatPenalty    *float64 `json:"repeat_penalty,omitempty"`
}

type OllamaConfig struct {
	Url     string        `json:"url"`
	Stream  bool          `json:"stream"`
	Model   string        `json:"model"`
	Options OllamaOptions `json:"options"`
}

type VllmConfig struct {
	Url               string         `json:"url"`
	Model             string         `json:"model"`
	Temperature       *float64       `json:"temperature,omitempty"`
	TopP              *float64       `json:"top_p,omitempty"`
	TopK              *int           `json:"top_k,omitempty"`
	MinP              *float64       `json:"min_p,omitempty"`
	Seed              *int           `json:"seed,omitempty"`
	Stop              []string       `json:"stop,omitempty"`
	PresencePenalty   *float64       `json:"presence_penalty,omitempty"`
	FrequencyPenalty  *float64       `json:"frequency_penalty,omitempty"`
	N                 *int           `json:"n,omitempty"`
	MinTokens         *int           `json:"min_tokens,omitempty"`
	IgnoreEOS         *bool          `json:"ignore_eos,omitempty"`
	GuidedJSON        map[string]any `json:"guided_json,omitempty"`
	GuidedRegex       *string        `json:"guided_regex,omitempty"`
	SkipSpecialTokens *bool          `json:"skip_special_tokens,omitempty"`
	MaxTokens         *int           `json:"max_tokens,omitempty"`
	RepetitionPenalty *float64       `json:"repetition_penalty,omitempty"`
}

type AgentConfig struct {
	LLMBackend   int          `json:"use_llm_backend"`
	OllamaConfig OllamaConfig `json:"ollama"`
	VllmConfig   VllmConfig   `json:"vllm"`
}

// The default config creates only for ollama
func CreateDefaultAgentConfig() AgentConfig {
	return AgentConfig{
		LLMBackend: 0,
		OllamaConfig: OllamaConfig{
			Url:    "http://localhost:11434/api/chat",
			Stream: false,
			Options: OllamaOptions{
				Temperature: ToPtr(0.7),
				NumCtx:      ToPtr(8192),
				TopP:        ToPtr(0.9),
			},
		},
	}
}

func ReadConfigFile(path string) (AgentConfig, error) {
	var conf AgentConfig
	src, err := os.ReadFile(path)
	if err != nil {
		return conf, err
	}

	err = json.Unmarshal(src, &conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}
