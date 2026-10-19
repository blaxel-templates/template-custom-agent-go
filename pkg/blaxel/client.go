package blaxel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"template-custom-agent-go/pkg/logger"

	blaxelsdk "github.com/blaxel-ai/sdk-go"
)

// Client represents a client for making requests to AI models
type Client struct {
	HTTPClient  *http.Client
	Workspace   string
	RunUrl      string
	ApiUrl      string
	Model       string
	Debug       bool
	Credentials blaxelsdk.Credentials
	McpManager  *MCPManager
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
	// BL_WORKSPACE first, then the current `bl login` context.
	workspace := blaxelsdk.GetDefaultWorkspace()
	// Picks the environment (BL_ENV or the workspace config) and applies BL_API_URL / BL_RUN_URL overrides.
	blaxelsdk.InitializeEnvironment(workspace)
	runUrl := blaxelsdk.GetRunURL()
	apiUrl := blaxelsdk.GetBaseURL()
	model := os.Getenv("BL_MODEL")
	if model == "" {
		model = "sandbox-openai"
	}
	debug := os.Getenv("BL_DEBUG")
	if debug == "" {
		debug = "false"
	}
	credentials := loadCredentials(workspace)
	if !credentials.IsValid() && workspace != "" {
		logger.Warningf("Invalid credentials for workspace %s", workspace)
		logger.Warningf("Please run `bl login %s` to fix it credentials.", workspace)
	}

	headers, err := credentials.AuthHeaders(context.Background(), workspace)
	if err != nil {
		logger.Fatalf("failed to get headers: %v", err)
	}

	// Initialize MCP Manager
	mcpManager := NewMCPManager(headers)

	// Configure MCP servers connected to
	serverNames := []string{"blaxel-search"}
	mcpServers := getMCPServersConfig(runUrl, workspace, serverNames)
	for _, serverConfig := range mcpServers {
		if err := mcpManager.AddServer(serverConfig); err != nil {
			logger.Warningf("Failed to add MCP server %s: %v", serverConfig.Name, err)
		}
	}

	return &Client{
		HTTPClient:  &http.Client{},
		Workspace:   workspace,
		Model:       model,
		Debug:       debug == "true",
		Credentials: credentials,
		RunUrl:      runUrl,
		ApiUrl:      apiUrl,
		McpManager:  mcpManager,
	}
}

// loadCredentials reads BL_API_KEY, then BL_CLIENT_CREDENTIALS, then the `bl login` config for the workspace.
func loadCredentials(workspace string) blaxelsdk.Credentials {
	if apiKey := os.Getenv("BL_API_KEY"); apiKey != "" {
		return blaxelsdk.Credentials{APIKey: apiKey}
	}
	if clientCredentials := os.Getenv("BL_CLIENT_CREDENTIALS"); clientCredentials != "" {
		return blaxelsdk.Credentials{ClientCredentials: clientCredentials}
	}
	credentials, err := blaxelsdk.LoadCredentials(workspace)
	if err != nil {
		logger.Warningf("Could not read Blaxel credentials: %v", err)
	}
	return credentials
}

// CreateChatCompletion sends a chat completion request
func (c *Client) CreateChatCompletion(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	ctx := context.Background()

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Fetch headers per request so expiring tokens get refreshed.
	headers, err := c.Credentials.AuthHeaders(ctx, c.Workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth headers: %w", err)
	}

	url := fmt.Sprintf("%s/%s/models/%s/v1/chat/completions", strings.TrimRight(c.RunUrl, "/"), c.Workspace, c.Model)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}
	if c.Debug {
		logger.Debugf("POST %s", url)
	}

	resp, err := c.HTTPClient.Do(httpReq)
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
