package apimerchant

import "github.com/gin-gonic/gin"

// Healthcheck checks health of the service.
func (s *Server) Healthcheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "OK",
	})
}
