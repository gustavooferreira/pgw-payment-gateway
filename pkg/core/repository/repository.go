package repository

import (
	"fmt"

	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/entities"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Database struct {
	connection *gorm.DB
}

func NewDatabase(host string, port int, username string, password string, dbname string) (*Database, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, port, dbname)

	dbconn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// TODO: Setup logger for gorm here

	// Check currency existance

	// Check if returns RecordNotFound error
	// var currencyResult entities.Currency
	// currency := entities.Currency{Name: "EUR"}
	// result := dbconn.Where(&currency).Take(&currencyResult)
	// if errors.Is(result.Error, gorm.ErrRecordNotFound) {
	// 	fmt.Println("not found")
	// } else if result.Error != nil {
	// 	fmt.Println("Some error")
	// } else {
	// 	fmt.Println("match found!")
	// }

	// auth := entities.Authorisation{ID: "aaa-bbb-ccc", State: entities.State{Name: "Authorised"},
	// 	Currency: entities.Currency{Name: "EUR"}, Amount: 10.50, MerchantName: "yolo Merchant"}

	// cc := entities.CreditCard{Number: 111, Name: "customer1", ExpiryMonth: 10, ExpiryYear: 2050, CVV: 123,
	// 	Authorisations: []entities.Authorisation{auth}}

	// result = dbconn.Create(&cc)
	// fmt.Println(result.Error)
	// fmt.Println(result.RowsAffected)

	db := Database{connection: dbconn}

	return &db, nil
}

func (db *Database) Close() error {
	sqlDB, err := db.connection.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

func (db *Database) HealthCheck() error {
	sqlDB, err := db.connection.DB()
	if err != nil {
		return err
	}

	err = sqlDB.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) GetAllAuthorisations() []entities.Authorisation {

	return make([]entities.Authorisation, 1, 1)
}

func (db *Database) GetAuthorisation() entities.Authorisation {

	return entities.Authorisation{}
}
