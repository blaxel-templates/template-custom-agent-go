package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// setupHealthRoutes sets up health check routes
func (r *Router) setupHealthRoutes(engine *gin.Engine) {
	health := engine.Group("/health")
	{
		health.GET("", r.healthCheck)
		health.GET("/ready", r.readinessCheck)
		health.GET("/live", r.livenessCheck)
	}
}

// healthCheck handles basic health check requests
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "template-custom-agent-go",
		"version": "1.0.0",
	})
}

// readinessCheck handles readiness probe requests
func (r *Router) readinessCheck(c *gin.Context) {
	// Check if MCP servers are available
	serverCount := r.blaxelClient.McpManager.GetServerCount()

	if serverCount == 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "no MCP servers available",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "ready",
		"mcp_servers": serverCount,
	})
}

// livenessCheck handles liveness probe requests
func (r *Router) livenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}
