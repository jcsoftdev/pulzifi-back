package middleware

import (
	"github.com/gin-gonic/gin"
)

// Health check endpoint
func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Service is healthy",
		})
	}
}
