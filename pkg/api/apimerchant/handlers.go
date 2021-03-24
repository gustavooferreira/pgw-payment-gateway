package apimerchant

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/api"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/api/middleware"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/entities"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/pprocessor"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/repository"
)

// AuthoriseTransaction handles authorisation of transactions.
func (s *Server) AuthoriseTransaction(c *gin.Context) {
	requestBody := struct {
		CreditCard struct {
			Name        string `json:"name" binding:"required"`
			Number      uint64 `json:"number" binding:"required"`
			ExpiryMonth uint   `json:"expiry_month" binding:"required"`
			ExpiryYear  uint   `json:"expiry_year" binding:"required"`
			CVV         uint   `json:"cvv" binding:"required"`
		} `json:"credit_card" binding:"required"`
		Currency string  `json:"currency" binding:"required"`
		Amount   float64 `json:"amount" binding:"required"`
	}{}

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		s.Logger.Info(fmt.Sprintf("error parsing body: %s", err.Error()))
		api.RespondWithError(c, 400, "error parsing body")
		return
	}

	responseBody := struct {
		AuthorisationID string  `json:"authorisation_id,omitempty"`
		Status          string  `json:"status"`
		ErrorMessage    string  `json:"error_message,omitempty"`
		Amount          float64 `json:"amount,omitempty"`
		Currency        string  `json:"currency,omitempty"`
	}{}

	// Validate credit card number
	if !core.LuhnValid(requestBody.CreditCard.Number) {
		errMessage := "credit card number provided does not pass Luhn check"
		s.Logger.Info(errMessage)
		api.RespondWithError(c, 400, errMessage)
		return
	}

	// Validate expiry date
	if !core.CardExpiryValid(requestBody.CreditCard.ExpiryYear, requestBody.CreditCard.ExpiryMonth) {
		errMessage := "credit card provided has expired"
		s.Logger.Info(errMessage)
		api.RespondWithError(c, 400, errMessage)
		return
	}

	// make external request to payment processor
	authReq := pprocessor.AuthorisationRequest{
		Currency: requestBody.Currency,
		Amount:   requestBody.Amount,
		CreditCard: pprocessor.CreditCard{
			Name:        requestBody.CreditCard.Name,
			Number:      requestBody.CreditCard.Number,
			ExpiryMonth: requestBody.CreditCard.ExpiryMonth,
			ExpiryYear:  requestBody.CreditCard.ExpiryYear,
			CVV:         requestBody.CreditCard.CVV,
		},
	}
	authID, ok := s.PProcessor.AuthorisePayment(authReq)
	if !ok {
		responseBody.Status = "fail"
		c.JSON(200, responseBody)
		return
	}

	// Get merchant_name
	merchantName := c.MustGet(middleware.AuthUserKey).(string)

	authRecord := entities.Authorisation{
		ID:           authID,
		State:        "Authorised",
		Currency:     requestBody.Currency,
		Amount:       requestBody.Amount,
		MerchantName: merchantName,
		CreditCard: &entities.CreditCard{
			Number:      requestBody.CreditCard.Number,
			Name:        requestBody.CreditCard.Name,
			ExpiryMonth: requestBody.CreditCard.ExpiryMonth,
			ExpiryYear:  requestBody.CreditCard.ExpiryYear,
			CVV:         requestBody.CreditCard.CVV,
		},
	}

	err = s.Repo.AddAuthorisation(authRecord)
	if e, ok := err.(*repository.DBServiceError); ok {
		if e.ValidationFail {
			s.Logger.Info(err.Error())
			api.RespondWithError(c, 400, err.Error())
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

	responseBody.Amount = requestBody.Amount
	responseBody.Currency = requestBody.Currency
	responseBody.Status = "success"
	responseBody.AuthorisationID = authID

	c.JSON(200, responseBody)
}

// CaptureTransaction handles capturing of transactions.
func (s *Server) CaptureTransaction(c *gin.Context) {
	requestBody := struct {
		AuthorisationID string  `json:"authorisation_id" binding:"required"`
		Amount          float64 `json:"amount" binding:"required"`
	}{}

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		s.Logger.Info(fmt.Sprintf("error parsing body: %s", err.Error()))
		api.RespondWithError(c, 400, "error parsing body")
		return
	}

	responseBody := struct {
		Status       string  `json:"status"`
		ErrorMessage string  `json:"error_message,omitempty"`
		Amount       float64 `json:"amount,omitempty"`
		Currency     string  `json:"currency,omitempty"`
	}{}

	// Get merchant_name
	merchantName := c.MustGet(middleware.AuthUserKey).(string)

	// Check if authID is in authorisations table
	authDetails, err := s.Repo.GetAuthorisationDetails(requestBody.AuthorisationID)
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

	// check merchant name match
	if authDetails.MerchantName != merchantName {
		api.RespondWithError(c, 403, "forbidden")
	}

	// check state is either "authorised" or "captured"
	if authDetails.State != "Authorised" && authDetails.State != "Captured" {
		api.RespondWithError(c, 400, fmt.Sprintf("payment has been %q", authDetails.State))
	}

	// Check we still can capture money (no limit yet)

	// Send request to payment processor

	// if successfull update DB with new transaction and state

	c.JSON(200, responseBody)
}

// RefundTransaction handles refunding of transactions.
func (s *Server) RefundTransaction(c *gin.Context) {
	requestBody := struct {
		AuthorisationID string  `json:"authorisation_id" binding:"required"`
		Amount          float64 `json:"amount" binding:"required"`
	}{}

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		s.Logger.Info(fmt.Sprintf("error parsing body: %s", err.Error()))
		api.RespondWithError(c, 400, "error parsing body")
		return
	}

	responseBody := struct {
		Status       string  `json:"status"`
		ErrorMessage string  `json:"error_message,omitempty"`
		Amount       float64 `json:"amount,omitempty"`
		Currency     string  `json:"currency,omitempty"`
	}{}

	// Check if authID is in authorisations table
	// Also check we are in the "captured" or "refunded" state

	// Check we still have money to refund

	// Send request to payment processor

	// if successfull update DB with new transaction

	c.JSON(200, responseBody)
}

// VoidTransaction handles voiding transactions.
func (s *Server) VoidTransaction(c *gin.Context) {
	requestBody := struct {
		AuthorisationID string `json:"authorisation_id" binding:"required"`
	}{}

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		s.Logger.Info(fmt.Sprintf("error parsing body: %s", err.Error()))
		api.RespondWithError(c, 400, "error parsing body")
		return
	}

	responseBody := struct {
		Status       string `json:"status"`
		ErrorMessage string `json:"error_message,omitempty"`
	}{}

	// Check if authID is in authorisations table
	// Also check we are in the "authorised" state

	// Send request to payment processor

	// if successful, update authorisation table with voided state

	c.JSON(200, responseBody)
}
