package repository

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Database struct {
	conn *gorm.DB
}

func NewDatabase(host string, port int, username string, password string, dbname string) (*Database, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, port, dbname)

	// dbconn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	dbconn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// create session
	dbconn = dbconn.Session(&gorm.Session{})
	dbconn = dbconn.Debug()

	// TODO: Setup logger for gorm here

	db := Database{conn: dbconn}

	return &db, nil
}

func (db *Database) Close() error {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

func (db *Database) HealthCheck() error {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return err
	}

	err = sqlDB.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) GetCurrencyID(currency string) (uint64, error) {
	var currencyResult Currency
	result := db.conn.Where(&Currency{Name: currency}).Take(&currencyResult)
	return currencyResult.ID, result.Error
}

func (db *Database) GetStateID(state string) (uint64, error) {
	var stateResult State
	result := db.conn.Where(&State{Name: state}).Take(&stateResult)
	return stateResult.ID, result.Error
}

func (db *Database) GetCreditCardDetails(number uint64) (CreditCard, error) {
	var creditcardResult CreditCard
	result := db.conn.Where(&CreditCard{Number: number}).Take(&creditcardResult)
	return creditcardResult, result.Error
}

func (db *Database) GetAuthorisationRecord(authID string) (Authorisation, error) {
	var authResult Authorisation
	result := db.conn.Preload("State").Preload("Currency").Where(&Authorisation{ID: authID}).Take(&authResult)
	return authResult, result.Error
}

func (db *Database) UpdateAuthorisationState(authID string, stateID uint64) error {
	result := db.conn.Model(&Authorisation{ID: authID}).Update("state_id", stateID)
	return result.Error
}

func (db *Database) FindAllAuthorisationRecords() ([]Authorisation, error) {
	var authResults []Authorisation
	result := db.conn.Preload("State").Preload("Currency").Find(&authResults)
	return authResults, result.Error
}

func (db *Database) InsertAuthorisationRecord(authRecord Authorisation) error {
	result := db.conn.Create(&authRecord)
	return result.Error
}

func (db *Database) InsertCreditCardRecord(ccRecord CreditCard) error {
	result := db.conn.Create(&ccRecord)
	return result.Error
}

func (db *Database) FindAllTransactionRecords(authID string) ([]Transaction, error) {
	var transactionResults []Transaction
	result := db.conn.Where(&Transaction{AuthorisationID: authID}).Find(&transactionResults)
	return transactionResults, result.Error
}

func (db *Database) InsertTransactionRecord(transRecord Transaction) error {
	result := db.conn.Create(&transRecord)
	return result.Error
}
