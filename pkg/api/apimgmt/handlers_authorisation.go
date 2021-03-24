package apimgmt

import (
	"github.com/gin-gonic/gin"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/api"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/repository"
)

// GetAuthorisations returns all authorisations from the database.
func (s *Server) GetAuthorisations(c *gin.Context) {

	authList, err := s.Repo.GetAllAuthorisations()
	if err != nil {
		s.Logger.Error(err.Error())
		api.RespondWithError(c, 500, "Internal error")
		return
	}

	c.JSON(200, authList)
}

// GetAuthorisation returns a detailed authorisation.
func (s *Server) GetAuthorisation(c *gin.Context) {
	authID := c.Param("authID")

	authDetails, err := s.Repo.GetAuthorisationDetails(authID)
	if e, ok := err.(*repository.DBServiceError); ok {
		if e.NotFound {
			api.RespondWithError(c, 404, err.Error())
			return
		}
		s.Logger.Error(err.Error())
		api.RespondWithError(c, 500, "Internal error")
		return
	} else if err != nil {
		s.Logger.Error(err.Error())
		api.RespondWithError(c, 500, "Internal error")
		return
	}

	c.JSON(200, authDetails)
}
