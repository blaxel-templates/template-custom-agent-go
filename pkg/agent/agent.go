package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"template-custom-agent-go/pkg/blaxel"
)

// Agent represents an AI agent with configurable model and tools
type Agent struct {
	name          string
	model         string
	tools         []blaxel.Tool
	blaxelClient  *blaxel.Client
	systemPrompt  string
	maxIterations int
	toolManager   *ToolManager
}

// Config holds configuration for creating an agent
type Config struct {
	Name          string
	Model         string
	SystemPrompt  string
	MaxIterations int
}

// NewAgent creates a new agent with the given configuration
func NewAgent(config Config, blaxelClient *blaxel.Client) *Agent {
	maxIterations := config.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 10
	}

	systemPrompt := config.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a helpful AI assistant. Use the available tools when needed to help answer user questions."
	}

	return &Agent{
		name:          config.Name,
		model:         config.Model,
		blaxelClient:  blaxelClient,
		systemPrompt:  systemPrompt,
		maxIterations: maxIterations,
		tools:         []blaxel.Tool{},
		toolManager:   NewToolManager(),
	}
}

// SetTools sets the tools available to the agent
func (a *Agent) SetTools(tools []blaxel.Tool) *Agent {
	a.tools = tools
	return a
}

// SetToolManager sets the tool manager for the agent
func (a *Agent) SetToolManager(tm *ToolManager) *Agent {
	a.toolManager = tm
	return a
}

// SetSystemPrompt sets the system prompt for the agent
func (a *Agent) SetSystemPrompt(prompt string) *Agent {
	a.systemPrompt = prompt
	return a
}

// SetMaxIterations sets the maximum number of iterations for the agent loop
func (a *Agent) SetMaxIterations(max int) *Agent {
	a.maxIterations = max
	return a
}

// Run executes the agent loop with the given user input
func (a *Agent) Run(ctx context.Context, userInput string) (*blaxel.ChatCompletionResponse, error) {
	// Initialize conversation
	messages := []blaxel.ChatMessage{
		{
			Role:    "system",
			Content: a.systemPrompt,
		},
		{
			Role:    "user",
			Content: userInput,
		},
	}

	// Run agent loop
	for iteration := 1; iteration <= a.maxIterations; iteration++ {
		// Send request to AI model
		req := blaxel.ChatCompletionRequest{
			Messages: messages,
			Tools:    a.tools,
		}

		fmt.Printf("Iteration %d: Sending request with %d tools\n", iteration, len(a.tools))
		if len(a.tools) > 0 {
			fmt.Printf("Tools being sent: %v\n", a.tools[0].Function.Name)
		}

		resp, err := a.blaxelClient.CreateChatCompletion(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get AI response (iteration %d): %w", iteration, err)
		}

		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("no response choices returned (iteration %d)", iteration)
		}

		assistantMessage := resp.Choices[0].Message
		fmt.Printf("Iteration %d: Assistant response has %d tool calls\n", iteration, len(assistantMessage.ToolCalls))
		messages = append(messages, assistantMessage)

		// Check if AI wants to use tools
		if len(assistantMessage.ToolCalls) > 0 {
			// Execute each tool call
			for _, toolCall := range assistantMessage.ToolCalls {
				toolResult, err := a.executeToolCall(ctx, toolCall)
				if err != nil {
					return nil, fmt.Errorf("failed to execute tool %s (iteration %d): %w",
						toolCall.Function.Name, iteration, err)
				}

				// Add tool result to conversation
				messages = append(messages, blaxel.ChatMessage{
					Role:       "tool",
					Content:    string(toolResult),
					ToolCallId: toolCall.Id,
				})
			}
			continue // Get next AI response with tool results
		}

		// No tool calls - this is the final response
		return resp, nil
	}

	// Max iterations reached
	return a.createMaxIterationsResponse(), nil
}

// executeToolCall executes a single tool call and returns the result
func (a *Agent) executeToolCall(ctx context.Context, toolCall blaxel.ToolCall) ([]byte, error) {
	// Parse parameters
	var params interface{}
	if toolCall.Function.Arguments != "" {
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
			return nil, fmt.Errorf("failed to parse tool arguments: %w", err)
		}
	}

	// Get the server for this tool
	serverName, exists := a.toolManager.GetServerForTool(toolCall.Function.Name)
	if !exists {
		return nil, fmt.Errorf("no server found for tool: %s", toolCall.Function.Name)
	}

	// Call the tool through the appropriate MCP server
	toolResult, err := a.blaxelClient.McpManager.CallTool(ctx, serverName, toolCall.Function.Name, params)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool %s: %w", toolCall.Function.Name, err)
	}
	content, err := json.Marshal(toolResult.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool result: %w", err)
	}
	return content, nil
}

// createMaxIterationsResponse creates a response when max iterations are reached
func (a *Agent) createMaxIterationsResponse() *blaxel.ChatCompletionResponse {
	return &blaxel.ChatCompletionResponse{
		ID:      fmt.Sprintf("agent-%s-%d", a.name, time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   a.model,
		Choices: []blaxel.Choice{
			{
				Index: 0,
				Message: blaxel.ChatMessage{
					Role:    "assistant",
					Content: "Maximum iterations reached. The agent may not have completed the task.",
				},
				FinishReason: "length",
			},
		},
		Usage: blaxel.UsageInfo{
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}
}

// GetName returns the agent's name
func (a *Agent) GetName() string {
	return a.name
}

// GetModel returns the agent's model
func (a *Agent) GetModel() string {
	return a.model
}

// GetToolsCount returns the number of tools available to the agent
func (a *Agent) GetToolsCount() int {
	return len(a.tools)
}
