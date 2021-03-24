package apimgmt

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// GetAuthorisations returns all authorisations from the database.
func (s *Server) GetAuthorisations(c *gin.Context) {

	// reply with all authorisations

	c.JSON(200, gin.H{"authKey": "authValue"})
}

// GetAuthorisation returns a detailed authorisation.
func (s *Server) GetAuthorisation(c *gin.Context) {
	authID := c.Param("authID")
	fmt.Println("Auth ID:", authID)
	// reply with a detailed authorisation

	c.JSON(200, gin.H{"authKey": "authValue"})
}
