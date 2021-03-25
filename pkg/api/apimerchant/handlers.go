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
		return
	}

	// check state is either "authorised" or "captured"
	if authDetails.State != "Authorised" && authDetails.State != "Captured" {
		api.RespondWithError(c, 400, fmt.Sprintf("cannot capture payment - payment has been '%s'", authDetails.State))
		return
	}

	// Check we can still capture money (haven't reached the limit yet)
	capturedSum := 0.0

	for _, transItem := range authDetails.Transaction {
		if transItem.Type == "Capture" {
			capturedSum += transItem.Amount
		}
	}

	if requestBody.Amount > authDetails.Amount-capturedSum {
		api.RespondWithError(c, 400, "cannot request more money than what was authorised")
		return
	}

	// make external request to payment processor
	captureReq := pprocessor.CaptureRequest{
		AuthorisationID: requestBody.AuthorisationID,
		Amount:          requestBody.Amount,
	}

	ok := s.PProcessor.CaptureTransaction(captureReq)
	if !ok {
		responseBody.Status = "fail"
		c.JSON(200, responseBody)
		return
	}

	// update DB with new transaction and state
	transItem := entities.Transaction{Type: "Capture", Amount: requestBody.Amount}
	err = s.Repo.AddTransaction(requestBody.AuthorisationID, transItem)
	if err != nil {
		s.Logger.Error(err.Error())
		api.RespondWithError(c, 500, "Internal error")
		return
	}

	responseBody.Amount = requestBody.Amount
	responseBody.Currency = authDetails.Currency
	responseBody.Status = "success"

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
		return
	}

	// check state is either "refunded" or "captured"
	if authDetails.State != "Refunded" && authDetails.State != "Captured" {
		api.RespondWithError(c, 400, fmt.Sprintf("cannot refund payment - payment has been '%s'", authDetails.State))
		return
	}

	// Check we can still refund money (haven't reached 0)
	capturedSum := 0.0

	for _, transItem := range authDetails.Transaction {
		if transItem.Type == "Capture" {
			capturedSum += transItem.Amount
		} else if transItem.Type == "Refund" {
			capturedSum -= transItem.Amount
		}
	}

	if requestBody.Amount > capturedSum {
		api.RespondWithError(c, 400, "cannot refund more money than what was captured")
		return
	}

	// make external request to payment processor
	refundReq := pprocessor.RefundRequest{
		AuthorisationID: requestBody.AuthorisationID,
		Amount:          requestBody.Amount,
	}

	ok := s.PProcessor.RefundTransaction(refundReq)
	if !ok {
		responseBody.Status = "fail"
		c.JSON(200, responseBody)
		return
	}

	// update DB with new transaction and state
	transItem := entities.Transaction{Type: "Refund", Amount: requestBody.Amount}
	err = s.Repo.AddTransaction(requestBody.AuthorisationID, transItem)
	if err != nil {
		s.Logger.Error(err.Error())
		api.RespondWithError(c, 500, "Internal error")
		return
	}

	responseBody.Amount = requestBody.Amount
	responseBody.Currency = authDetails.Currency
	responseBody.Status = "success"

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
		return
	}

	// check state is "authorised"
	if authDetails.State != "Authorised" {
		api.RespondWithError(c, 400, fmt.Sprintf("cannot void payment - payment has been '%s'", authDetails.State))
		return
	}

	// make external request to payment processor
	voidReq := pprocessor.VoidRequest{
		AuthorisationID: requestBody.AuthorisationID,
	}

	ok := s.PProcessor.VoidPayment(voidReq)
	if !ok {
		responseBody.Status = "fail"
		c.JSON(200, responseBody)
		return
	}

	// update DB with new transaction and state
	err = s.Repo.UpdateAuthorisationState(requestBody.AuthorisationID, "Voided")
	if err != nil {
		s.Logger.Error(err.Error())
		api.RespondWithError(c, 500, "Internal error")
		return
	}

	responseBody.Status = "success"

	c.JSON(200, responseBody)
}
