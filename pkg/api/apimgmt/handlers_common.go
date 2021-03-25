package apimgmt

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// Healthcheck checks health of the service.
func (s *Server) Healthcheck(c *gin.Context) {
	err := s.Repo.HealthCheck()
	if err != nil {
		s.Logger.Error(fmt.Sprintf("database health check error: %s", err.Error()))
		c.JSON(500, gin.H{"status": "FAIL"})
        return
	}

	c.JSON(200, gin.H{"status": "OK"})
}
