package core

import (
	"context"

	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/entities"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/pprocessor"
)

// Repository represents a database holding the data
type Repository interface {
	HealthCheck() error
	Close() error
	CurrencyExists(currency string) (bool, error)
	AddAuthorisation(entities.Authorisation) error

	GetAllAuthorisations() ([]entities.Authorisation, error)
	GetAuthorisationDetails(authID string) (entities.Authorisation, error)
}

// PaymentProcessor represents a payment processor service
type PaymentProcessor interface {
	AuthorisePayment(pprocessor.AuthorisationRequest) (authID string, success bool)
	CaptureTransaction(pprocessor.CaptureRequest) (success bool)
	RefundTransaction(pprocessor.RefundRequest) (success bool)
	VoidPayment(pprocessor.VoidRequest) (success bool)
}

// ShutDowner represents anything that can be shutdown like an HTTP server.
type ShutDowner interface {
	ShutDown(ctx context.Context) error
}
