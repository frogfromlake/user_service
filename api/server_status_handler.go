package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// readinessCheck is a simple endpoint to check if the server is ready to handle requests.
func (server *Server) readinessCheck(c *gin.Context) {
	ctx := c.Request.Context()
	if err := server.store.Ping(ctx, 5*time.Second); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Database not ready",
		})
		return
	}

	// Add more checks for other dependencies if needed...

	// If all checks passed, return a 200 OK response.
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
