package blaxel

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"template-custom-agent-go/pkg/logger"

	blaxelMCP "github.com/blaxel-ai/toolkit/sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPServerConfig represents configuration for a single MCP server
type MCPServerConfig struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// MCPManager manages multiple MCP servers
type MCPManager struct {
	servers map[string]*blaxelMCP.MCPClient
	headers map[string]string
}

// ToolWithServer represents a tool with its associated server
type ToolWithServer struct {
	Tool       *mcp.Tool
	ServerName string
}

// NewMCPManager creates a new MCP manager
func NewMCPManager(headers map[string]string) *MCPManager {
	return &MCPManager{
		servers: make(map[string]*blaxelMCP.MCPClient),
		headers: headers,
	}
}

// AddServer adds a new MCP server to the manager
func (m *MCPManager) AddServer(config MCPServerConfig) error {
	client, err := blaxelMCP.NewMCPClient(config.URL, m.headers)
	if err != nil {
		return fmt.Errorf("failed to create MCP client for %s: %w", config.Name, err)
	}

	m.servers[config.Name] = client
	logger.Debugf("Added MCP server: %s at %s", config.Name, config.URL)
	return nil
}

// ListAllTools aggregates tools from all connected MCP servers
func (m *MCPManager) ListAllTools(ctx context.Context) ([]ToolWithServer, error) {
	var allTools []ToolWithServer

	for serverName, client := range m.servers {
		tools, err := client.ListTools(ctx)
		if err != nil {
			logger.Warningf("Failed to get tools from server %s: %v", serverName, err)
			continue
		}

		for _, tool := range tools.Tools {
			allTools = append(allTools, ToolWithServer{
				Tool:       tool,
				ServerName: serverName,
			})
		}
	}

	return allTools, nil
}

// CallTool routes a tool call to the appropriate MCP server
func (m *MCPManager) CallTool(ctx context.Context, serverName, toolName string, params interface{}) (*mcp.CallToolResult, error) {
	client, exists := m.servers[serverName]
	if !exists {
		return nil, fmt.Errorf("MCP server %s not found", serverName)
	}

	return client.CallTool(ctx, toolName, params)
}

// GetServerNames returns a list of all connected server names
func (m *MCPManager) GetServerNames() []string {
	var names []string
	for name := range m.servers {
		names = append(names, name)
	}
	return names
}

// GetServerCount returns the number of connected servers
func (m *MCPManager) GetServerCount() int {
	return len(m.servers)
}

// Close closes all MCP server connections
func (m *MCPManager) Close() error {
	var lastErr error
	for name, client := range m.servers {
		if err := client.Close(); err != nil {
			logger.Errorf("Error closing MCP server %s: %v", name, err)
			lastErr = err
		}
	}
	return lastErr
}

// getMCPServersConfig returns MCP server configurations
// Can be extended to read from config file or environment variables
func getMCPServersConfig(runUrl, workspace string, serverNames []string) []MCPServerConfig {
	// Default configuration - can be extended to read from config file
	servers := []MCPServerConfig{}

	for _, serverName := range serverNames {
		servers = append(servers, MCPServerConfig{
			Name: serverName,
			URL:  fmt.Sprintf("%s/%s/functions/%s", runUrl, workspace, serverName),
		})
	}

	return servers
}

// LoadMCPServersFromConfig loads MCP server configurations from a config file
func LoadMCPServersFromConfig(configPath string) ([]MCPServerConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config struct {
		MCPServers []MCPServerConfig `json:"mcp_servers"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config.MCPServers, nil
}
