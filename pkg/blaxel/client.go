package blaxel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/blaxel-ai/toolkit/sdk"
)

// Client represents a client for making requests to AI models
type Client struct {
	BlaxelClient *sdk.ClientWithResponses
	Workspace    string
	RunUrl       string
	ApiUrl       string
	Model        string
	Debug        bool
	AuthProvider sdk.AuthProvider
	McpManager   *MCPManager
}

// ChatCompletionRequest represents the request body for chat completions
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
	Tools       []Tool        `json:"tools,omitempty"`
	ToolChoice  interface{}   `json:"tool_choice,omitempty"`
}

// Tool represents a tool that can be called by the AI
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function represents a function definition
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall represents a tool call made by the AI
type ToolCall struct {
	Id       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction represents the function part of a tool call
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatMessage represents a single message in the conversation
type ChatMessage struct {
	Role       string     `json:"role"` // "system", "user", "assistant", "tool"
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallId string     `json:"tool_call_id,omitempty"`
}

// ChatCompletionResponse represents the response from the chat completions API
type ChatCompletionResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   UsageInfo `json:"usage"`
}

// Choice represents a single completion choice
type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// UsageInfo represents token usage information
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// NewClient creates a new Blaxel client
func NewClient() *Client {
	workspace := os.Getenv("BL_WORKSPACE")
	if workspace == "" {
		workspace = sdk.CurrentContext().Workspace
	}
	runUrl := os.Getenv("BL_RUN_URL")
	if runUrl == "" {
		runUrl = "https://run.blaxel.ai"
	}
	apiUrl := os.Getenv("BL_API_URL")
	if apiUrl == "" {
		apiUrl = "https://api.blaxel.ai/v0"
	}
	model := os.Getenv("BL_MODEL")
	if model == "" {
		model = "sandbox-openai"
	}
	debug := os.Getenv("BL_DEBUG")
	if debug == "" {
		debug = "false"
	}
	var credentials sdk.Credentials
	if os.Getenv("BL_CLIENT_CREDENTIALS") != "" {
		credentials = sdk.Credentials{
			ClientCredentials: os.Getenv("BL_CLIENT_CREDENTIALS"),
		}
	} else {
		credentials = sdk.LoadCredentials(workspace)
	}
	if !credentials.IsValid() && workspace != "" {
		fmt.Printf("Invalid credentials for workspace %s\n", workspace)
		fmt.Printf("Please run `bl login %s` to fix it credentials.\n", workspace)
	}
	c, err := sdk.NewClientWithCredentials(
		sdk.RunClientWithCredentials{
			ApiURL:      apiUrl,
			RunURL:      runUrl,
			Credentials: credentials,
			Workspace:   workspace,
		},
	)
	if err != nil {
		log.Fatalf("Error creating Blaxel client: %v\n", err)
	}
	authProvider := sdk.GetAuthProvider(credentials, workspace, apiUrl)

	headers, err := authProvider.GetHeaders()
	if err != nil {
		log.Fatalf("failed to get headers: %v", err)
	}

	// Initialize MCP Manager
	mcpManager := NewMCPManager(headers)

	// Configure MCP servers from environment or use default blaxel-search
	serverNames := []string{"blaxel-search", ""}
	mcpServers := getMCPServersConfig(runUrl, workspace, serverNames)
	for _, serverConfig := range mcpServers {
		if err := mcpManager.AddServer(serverConfig); err != nil {
			log.Printf("Warning: Failed to add MCP server %s: %v", serverConfig.Name, err)
		}
	}

	return &Client{
		BlaxelClient: c,
		Workspace:    workspace,
		Model:        model,
		Debug:        debug == "true",
		AuthProvider: authProvider,
		RunUrl:       runUrl,
		ApiUrl:       apiUrl,
		McpManager:   mcpManager,
	}
}

// CreateChatCompletion sends a chat completion request
func (c *Client) CreateChatCompletion(req ChatCompletionRequest) (*ChatCompletionResponse, error) {

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.BlaxelClient.Run(
		context.Background(),
		c.Workspace,
		"model",
		c.Model,
		"POST",
		"/v1/chat/completions",
		map[string]string{},
		[]string{},
		string(jsonData),
		c.Debug,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API error: %s", errorResp.Error.Message)
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &chatResp, nil
}

// CreateSimpleCompletion is a helper function for simple text completions
func (c *Client) CreateSimpleCompletion(prompt string) (string, error) {
	req := ChatCompletionRequest{
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	resp, err := c.CreateChatCompletion(req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned in response")
	}

	return resp.Choices[0].Message.Content, nil
}
