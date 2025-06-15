package router

import (
	"fmt"
	"net/http"

	"template-custom-agent-go/pkg/blaxel"

	"github.com/gin-gonic/gin"
)

// setupChatRoutes sets up chat-related routes
func (r *Router) setupChatRoutes(engine *gin.Engine) {
	// OpenAI-compatible endpoint
	v1 := engine.Group("/v1")
	{
		v1.POST("/chat/completions", r.chatCompletions)
	}

	// Simple chat endpoint
	engine.POST("/chat", r.simpleChat)
}

// chatCompletions handles OpenAI-compatible chat completion requests
func (r *Router) chatCompletions(c *gin.Context) {
	var req blaxel.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(fmt.Errorf("invalid request format: %w", err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	resp, err := r.blaxelClient.CreateChatCompletion(req)
	if err != nil {
		c.Error(fmt.Errorf("failed to get AI response: %w", err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// simpleChat handles simple chat requests
func (r *Router) simpleChat(c *gin.Context) {
	var request struct {
		Message string `json:"message" binding:"required"`
		Model   string `json:"model"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(fmt.Errorf("invalid request: %w", err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	response, err := r.blaxelClient.CreateSimpleCompletion(request.Message)
	if err != nil {
		c.Error(fmt.Errorf("failed to get AI response: %w", err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": response,
		"model":    request.Model,
	})
}
