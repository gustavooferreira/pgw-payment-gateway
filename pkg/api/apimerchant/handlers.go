package apimerchant

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/api"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/api/middleware"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core"
)

// AuthoriseTransaction handles authorisation of transactions.
func (s *Server) AuthoriseTransaction(c *gin.Context) {
	requestBody := struct {
		CreditCard struct {
			Name        string `json:"name" binding:"required"`
			Number      int64  `json:"number" binding:"required"`
			ExpiryMonth int    `json:"expiry_month" binding:"required"`
			ExpiryYear  int    `json:"expiry_year" binding:"required"`
			CVV         int    `json:"cvv" binding:"required"`
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
		responseBody.Status = "fail"
		responseBody.ErrorMessage = errMessage
		c.JSON(400, responseBody)
		return
	}

	// Validate expiry date
	if !core.CardExpiryValid(requestBody.CreditCard.ExpiryYear, requestBody.CreditCard.ExpiryMonth) {
		errMessage := "credit card provided has expired"
		s.Logger.Info(errMessage)
		responseBody.Status = "fail"
		responseBody.ErrorMessage = errMessage
		c.JSON(400, responseBody)
		return
	}

	// validate currency

	// make external request to payment processor

	// Get merchant_name
	merchantName := c.MustGet(middleware.AuthUserKey).(string)

	_ = merchantName

	// Store in DB

	responseBody.Amount = requestBody.Amount
	responseBody.Currency = requestBody.Currency
	responseBody.Status = "success"
	responseBody.AuthorisationID = "aaa-bbb-ccc"

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

	c.JSON(200, responseBody)
}
