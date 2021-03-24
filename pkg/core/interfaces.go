package core

import (
	"context"

	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/entities"
)

// Repository represents a database holding the data
type Repository interface {
	GetAllAuthorisations() []entities.Authorisation
	GetAuthorisation() entities.Authorisation
	HealthCheck() error
}

// ShutDowner represents anything that can be shutdown like an HTTP server.
type ShutDowner interface {
	ShutDown(ctx context.Context) error
}
