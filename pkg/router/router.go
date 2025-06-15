package router

import (
	"net/http"

	"template-custom-agent-go/pkg/blaxel"
	"template-custom-agent-go/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// Router holds the dependencies needed for all routes
type Router struct {
	blaxelClient *blaxel.Client
}

// NewRouter creates a new router with dependencies
func NewRouter(blaxelClient *blaxel.Client) *Router {
	return &Router{
		blaxelClient: blaxelClient,
	}
}

// SetupRoutes configures all routes for the application
func (r *Router) SetupRoutes() *gin.Engine {
	// Create a Gin router without default middleware
	engine := gin.New()

	// Add custom middleware stack
	engine.Use(middleware.LoggingMiddleware())        // Custom logging
	engine.Use(middleware.CustomRecoveryMiddleware()) // Custom panic recovery
	engine.Use(middleware.ErrorHandlerMiddleware())   // Custom error handling

	// Setup all route groups
	r.setupHealthRoutes(engine)
	r.setupToolRoutes(engine)
	r.setupAgentRoutes(engine)
	r.setupChatRoutes(engine)
	r.setupRootRoutes(engine)

	return engine
}

// setupRootRoutes sets up root and documentation routes
func (r *Router) setupRootRoutes(engine *gin.Engine) {
	engine.GET("/", r.rootEndpoint)
}

// rootEndpoint handles root endpoint requests
func (r *Router) rootEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to the Template Custom Agent Go",
		"version": "1.0.0",
		"endpoints": gin.H{
			"health": []string{
				"GET /health - Basic health check",
				"GET /health/ready - Readiness probe",
				"GET /health/live - Liveness probe",
			},
			"tools": []string{
				"GET /tools - List all tools from all MCP servers",
				"GET /tools/servers - List all MCP servers",
				"GET /tools/servers/:server/tools - List tools from specific server",
			},
			"agent": []string{
				"POST /agent - Run agent with tool calling",
				"POST /agent/run - Alternative agent endpoint",
			},
			"chat": []string{
				"POST /v1/chat/completions - OpenAI-compatible chat completions",
				"POST /chat - Simple chat interface",
			},
		},
		"features": []string{
			"Multi-MCP server support",
			"OpenAI-compatible API",
			"Tool calling and routing",
			"Health monitoring",
		},
	})
}
