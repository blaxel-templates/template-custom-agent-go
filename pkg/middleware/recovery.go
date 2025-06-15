package middleware

import (
	"net/http"
	"runtime/debug"
	"template-custom-agent-go/pkg/logger"
	"template-custom-agent-go/pkg/models"
	"time"

	"github.com/gin-gonic/gin"
)

// CustomRecoveryMiddleware handles panics and prevents server crashes
func CustomRecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Log the panic with stack trace
		logger.Errorf("PANIC RECOVERED: %v\n%s", recovered, debug.Stack())

		// Create standardized error response
		errorResp := models.ErrorResponse{
			Error:     "Internal server error - panic recovered",
			Code:      http.StatusInternalServerError,
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
		}

		// Return error response and abort further processing
		c.JSON(http.StatusInternalServerError, errorResp)
		c.Abort()
	})
}
