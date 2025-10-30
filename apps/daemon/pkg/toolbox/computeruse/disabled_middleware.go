package computeruse

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// computerUseDisabledMiddleware returns a middleware that handles requests when computer-use is disabled
func ComputerUseDisabledMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message":  "Computer-use functionality is not available",
			"details":  "The computer-use plugin failed to initialize due to missing dependencies in the runtime environment.",
			"solution": "Install the required X11 dependencies (x11-apps, xvfb, etc.) to enable computer-use functionality. Check the daemon logs for specific error details.",
		})
		c.Abort()
	}
}
