package router

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// setupToolRoutes sets up tool-related routes
func (r *Router) setupToolRoutes(engine *gin.Engine) {
	tools := engine.Group("/tools")
	{
		tools.GET("", r.listTools)
		tools.GET("/servers", r.listMCPServers)
		tools.GET("/servers/:server/tools", r.listServerTools)
	}
}

// listTools handles tool listing requests from all servers
func (r *Router) listTools(c *gin.Context) {
	tools, err := r.blaxelClient.McpManager.ListAllTools(c)
	if err != nil {
		c.Error(fmt.Errorf("failed to list tools: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tools":       tools,
		"total_count": len(tools),
	})
}

// listMCPServers handles MCP server listing requests
func (r *Router) listMCPServers(c *gin.Context) {
	serverNames := r.blaxelClient.McpManager.GetServerNames()
	serverCount := r.blaxelClient.McpManager.GetServerCount()

	c.JSON(http.StatusOK, gin.H{
		"servers": serverNames,
		"count":   serverCount,
	})
}

// listServerTools handles tool listing requests for a specific server
func (r *Router) listServerTools(c *gin.Context) {
	serverName := c.Param("server")

	// Get all tools and filter by server
	allTools, err := r.blaxelClient.McpManager.ListAllTools(c)
	if err != nil {
		c.Error(fmt.Errorf("failed to list tools: %w", err))
		return
	}

	var serverTools []interface{}
	for _, toolWithServer := range allTools {
		if toolWithServer.ServerName == serverName {
			serverTools = append(serverTools, toolWithServer.Tool)
		}
	}

	if len(serverTools) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":  "server not found or has no tools",
			"server": serverName,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"server": serverName,
		"tools":  serverTools,
		"count":  len(serverTools),
	})
}
