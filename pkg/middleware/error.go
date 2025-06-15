package middleware

import (
	"log"
	"net/http"
	"template-custom-agent-go/pkg/models"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware provides consistent error handling across all endpoints
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Process the request
		c.Next()

		// Check if there are any errors after processing
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()

			// Log the error
			log.Printf("Request error: %v, Path: %s, Method: %s", err.Error(), c.Request.URL.Path, c.Request.Method)

			// Determine status code if not already set
			statusCode := c.Writer.Status()
			if statusCode == http.StatusOK {
				statusCode = http.StatusInternalServerError
			}

			// Create standardized error response
			errorResp := models.ErrorResponse{
				Error:     err.Error(),
				Code:      statusCode,
				Timestamp: time.Now(),
				Path:      c.Request.URL.Path,
			}

			// Only send response if not already sent
			if !c.Writer.Written() {
				c.JSON(statusCode, errorResp)
			}
		}
	})
}
