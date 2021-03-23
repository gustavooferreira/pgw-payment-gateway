package api

import (
	"github.com/gin-gonic/gin"
)

// NoRoute provides a generic handler for unmatched routes.
func NoRoute(c *gin.Context) {
	c.JSON(404, gin.H{"message": "no route found"})
}

// RespondWithError is a helper function to return an error according to the API specification.
func RespondWithError(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, gin.H{"message": message})
}
