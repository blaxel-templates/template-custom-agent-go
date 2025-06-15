package router

import (
	"fmt"
	"net/http"
	"strings"

	"template-custom-agent-go/pkg/agent"

	"github.com/gin-gonic/gin"
)

// setupAgentRoutes sets up agent-related routes
func (r *Router) setupAgentRoutes(engine *gin.Engine) {
	agents := engine.Group("/agent")
	{
		agents.POST("", r.runAgent)
		agents.POST("/run", r.runAgent) // Alternative endpoint
	}
}

// runAgent handles agent execution requests
func (r *Router) runAgent(c *gin.Context) {
	var request struct {
		Inputs        string `json:"inputs" binding:"required"`
		MaxIterations int    `json:"max_iterations,omitempty"`
		Model         string `json:"model,omitempty"`
		SystemPrompt  string `json:"system_prompt,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(fmt.Errorf("invalid request: %w", err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Set defaults
	model := request.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	systemPrompt := request.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant that can answer questions and help with tasks."
	}

	// Create agent with configuration
	agentConfig := agent.Config{
		Name:          "demo-agent",
		MaxIterations: request.MaxIterations,
		Model:         model,
		SystemPrompt:  systemPrompt,
	}

	demoAgent := agent.NewAgent(agentConfig, r.blaxelClient)

	// Get and set available tools
	mcpTools, err := r.blaxelClient.McpManager.ListAllTools(c)
	if err != nil {
		c.Error(fmt.Errorf("failed to get tools: %w", err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	toolManager := agent.NewToolManager()
	tools := toolManager.ConvertMCPToolsToOpenAI(mcpTools)

	toolNames := []string{}
	for _, tool := range tools {
		toolNames = append(toolNames, tool.Function.Name)
	}

	// Set both tools and tool manager on the agent
	demoAgent.SetTools(tools)
	demoAgent.SetToolManager(toolManager)
	fmt.Printf("Agent configured with %s tools\n", strings.Join(toolNames, ", "))

	// Run the agent
	response, err := demoAgent.Run(c, request.Inputs)
	if err != nil {
		c.Error(fmt.Errorf("agent execution failed: %w", err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}
