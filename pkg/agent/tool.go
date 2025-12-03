package agent

import (
	"encoding/json"

	"template-custom-agent-go/pkg/blaxel"
)

// ToolManager handles conversion between MCP tools and OpenAI tools
type ToolManager struct {
	// Map to track which server each tool belongs to
	toolServerMap map[string]string
}

// NewToolManager creates a new tool manager
func NewToolManager() *ToolManager {
	return &ToolManager{
		toolServerMap: make(map[string]string),
	}
}

// ConvertMCPToolsToOpenAI converts MCP tools to OpenAI format and tracks server associations
func (tm *ToolManager) ConvertMCPToolsToOpenAI(mcpToolsWithServer []blaxel.ToolWithServer) []blaxel.Tool {
	var openAITools []blaxel.Tool

	// Clear previous mappings
	tm.toolServerMap = make(map[string]string)

	for _, toolWithServer := range mcpToolsWithServer {
		mcpTool := toolWithServer.Tool
		serverName := toolWithServer.ServerName

		// Store server association
		tm.toolServerMap[mcpTool.Name] = serverName

		// Handle optional description
		description := mcpTool.Description

		// Convert to OpenAI format
		openAITool := blaxel.Tool{
			Type: "function",
			Function: blaxel.Function{
				Name:        mcpTool.Name,
				Description: description,
				Parameters:  convertParameters(mcpTool.InputSchema),
			},
		}

		openAITools = append(openAITools, openAITool)
	}

	return openAITools
}

// GetServerForTool returns the server name for a given tool
func (tm *ToolManager) GetServerForTool(toolName string) (string, bool) {
	serverName, exists := tm.toolServerMap[toolName]
	return serverName, exists
}

// convertParameters converts MCP input schema to OpenAI parameters format
func convertParameters(inputSchema interface{}) map[string]interface{} {
	// Convert to JSON and back to get a clean map[string]interface{}
	jsonBytes, err := json.Marshal(inputSchema)
	if err != nil {
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}

	return result
}
