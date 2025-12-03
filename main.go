package main

import (
	"os"
	"template-custom-agent-go/pkg/blaxel"
	"template-custom-agent-go/pkg/logger"
	"template-custom-agent-go/pkg/router"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	// Initialize Blaxel client
	bl := blaxel.NewClient()

	// Create router with dependencies
	r := router.NewRouter(bl)

	// Setup all routes
	engine := r.SetupRoutes()

	// Get host from environment variable or use default
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	// Start server on the specified port
	logger.Infof("Starting server on port %s", port)
	if err := engine.Run(host + ":" + port); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
