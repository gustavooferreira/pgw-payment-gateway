package repository

import (
	"errors"
	"fmt"

	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/entities"
	"gorm.io/gorm"
)

type DBServiceError struct {
	Msg            string
	ValidationFail bool
	NotFound       bool
	Err            error
}

func (e *DBServiceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Msg, e.Err.Error())
	}
	return e.Msg
}
func (e *DBServiceError) Unwrap() error {
	return e.Err
}

type DatabaseService struct {
	Database *Database
}

func NewDatabaseService(host string, port int, username string, password string, dbname string) (dbs *DatabaseService, err error) {
	dbs = &DatabaseService{}
	dbs.Database, err = NewDatabase(host, port, username, password, dbname)
	if err != nil {
		return nil, err
	}

	return dbs, nil
}

func (dbs *DatabaseService) Close() error {
	return dbs.Database.Close()
}

func (dbs *DatabaseService) HealthCheck() error {
	return dbs.Database.HealthCheck()
}

func (dbs *DatabaseService) CurrencyExists(currency string) (bool, error) {
	_, err := dbs.Database.GetCurrencyID(currency)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil // Not found
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func (dbs *DatabaseService) AddAuthorisation(auth entities.Authorisation) error {
	// Check currency is supported
	currencyID, err := dbs.Database.GetCurrencyID(auth.Currency)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &DBServiceError{Msg: "currency provided not supported", ValidationFail: true, Err: gorm.ErrRecordNotFound}
	} else if err != nil {
		return &DBServiceError{Msg: "database error", Err: err}
	}

	// get stateID
	stateID, err := dbs.Database.GetStateID(auth.State)
	if err != nil {
		return &DBServiceError{Msg: "database error", Err: err}
	}

	// Check whether credit card exists, if not, create it
	creditCard, err := dbs.Database.GetCreditCardDetails(auth.CreditCard.Number)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create credit card
		creditCard = CreditCard{
			Number:      auth.CreditCard.Number,
			Name:        auth.CreditCard.Name,
			ExpiryMonth: auth.CreditCard.ExpiryMonth,
			ExpiryYear:  auth.CreditCard.ExpiryYear,
			CVV:         auth.CreditCard.CVV,
		}
		err = dbs.Database.InsertCreditCardRecord(creditCard)
		if err != nil {
			return &DBServiceError{Msg: "database error", Err: err}
		}
	} else if err != nil {
		return &DBServiceError{Msg: "database error", Err: err}
	}

	// check if authID already exists
	_, err = dbs.Database.GetAuthorisationRecord(auth.ID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return &DBServiceError{Msg: "database error", Err: err}
		}
	} else {
		// error! record already exists
		return &DBServiceError{Msg: "authorisation ID already exists in the database", ValidationFail: false}
	}

	// Create authorisation record
	authRecord := Authorisation{
		ID:               auth.ID,
		StateID:          stateID,
		CurrencyID:       currencyID,
		Amount:           auth.Amount,
		MerchantName:     auth.MerchantName,
		CreditCardNumber: creditCard.Number,
	}

	err = dbs.Database.InsertAuthorisationRecord(authRecord)
	if err != nil {
		return &DBServiceError{Msg: "database error", Err: err}
	}

	return nil
}

func (dbs *DatabaseService) GetAllAuthorisations() ([]entities.Authorisation, error) {
	authorisations, err := dbs.Database.FindAllAuthorisationRecords()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []entities.Authorisation{}, nil
	} else if err != nil {
		return nil, &DBServiceError{Msg: "database error", Err: err}
	}

	authList := make([]entities.Authorisation, 0, len(authorisations))

	for _, authRecord := range authorisations {
		authItem := entities.Authorisation{
			ID:           authRecord.ID,
			State:        authRecord.State.Name,
			Currency:     authRecord.Currency.Name,
			Amount:       authRecord.Amount,
			MerchantName: authRecord.MerchantName,
		}

		authList = append(authList, authItem)
	}

	return authList, nil
}

func (dbs *DatabaseService) GetAuthorisationDetails(authID string) (authItem entities.Authorisation, err error) {

	// check if authID exists
	authRecord, err := dbs.Database.GetAuthorisationRecord(authID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return authItem, &DBServiceError{Msg: "authorisation record not found", NotFound: true}
	} else if err != nil {
		return authItem, &DBServiceError{Msg: "database error", Err: err}
	}

	authItem = entities.Authorisation{
		ID:           authRecord.ID,
		State:        authRecord.State.Name,
		Currency:     authRecord.Currency.Name,
		Amount:       authRecord.Amount,
		MerchantName: authRecord.MerchantName,
	}

	// get credit card information
	creditCardRecord, err := dbs.Database.GetCreditCardDetails(authRecord.CreditCardNumber)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return authItem, &DBServiceError{Msg: "credit card record not found", NotFound: true}
	} else if err != nil {
		return authItem, &DBServiceError{Msg: "database error", Err: err}
	}

	authItem.CreditCard = &entities.CreditCard{
		Number:      creditCardRecord.Number,
		Name:        creditCardRecord.Name,
		ExpiryMonth: creditCardRecord.ExpiryMonth,
		ExpiryYear:  creditCardRecord.ExpiryYear,
		CVV:         creditCardRecord.CVV,
	}

	// get all transactions associated with this authorisation
	transactionRecords, err := dbs.Database.FindAllTransactionRecords(authID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return authItem, nil
	} else if err != nil {
		return authItem, &DBServiceError{Msg: "database error", Err: err}
	}

	transactionsList := make([]entities.Transaction, 0, len(transactionRecords))

	for _, transRecord := range transactionRecords {
		transItem := entities.Transaction{
			ID:     transRecord.AuthorisationID,
			Type:   transRecord.Type,
			Amount: transRecord.Amount,
		}

		transactionsList = append(transactionsList, transItem)
	}

	authItem.Transaction = transactionsList

	return authItem, nil
}

func (dbs *DatabaseService) AddTransaction(authID string, transaction entities.Transaction) error {
	state := "Captured"
	if transaction.Type == "Refund" {
		state = "Refunded"
	}

	stateID, err := dbs.Database.GetStateID(state)
	if err != nil {
		return &DBServiceError{Msg: "database error", Err: err}
	}

	err = dbs.Database.UpdateAuthorisationState(authID, stateID)
	if err != nil {
		return &DBServiceError{Msg: "database error", Err: err}
	}

	transRecord := Transaction{
		Type:            transaction.Type,
		Amount:          transaction.Amount,
		AuthorisationID: authID,
	}

	err = dbs.Database.InsertTransactionRecord(transRecord)
	if err != nil {
		return &DBServiceError{Msg: "database error", Err: err}
	}

	return nil
}

func (dbs *DatabaseService) UpdateAuthorisationState(authID string, state string) error {
	stateID, err := dbs.Database.GetStateID(state)
	if err != nil {
		return &DBServiceError{Msg: "database error", Err: err}
	}

	err = dbs.Database.UpdateAuthorisationState(authID, stateID)
	if err != nil {
		return &DBServiceError{Msg: "database error", Err: err}
	}

	return nil
}
