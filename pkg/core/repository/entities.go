package repository

type CreditCard struct {
	Number         uint64          `gorm:"primaryKey;not null"`
	Name           string          `gorm:"type:varchar(50);not null"`
	ExpiryMonth    uint            `gorm:"not null"`
	ExpiryYear     uint            `gorm:"not null"`
	CVV            uint            `gorm:"not null"`
	Authorisations []Authorisation `gorm:"foreignKey:CreditCardNumber;not null"`
}

type Authorisation struct {
	ID               string `gorm:"primaryKey;type:varchar(50);not null"`
	State            State
	StateID          uint64 `gorm:"not null"` // Foreign Key
	Currency         Currency
	CurrencyID       uint64        `gorm:"not null"` // Foreign Key
	Amount           float64       `gorm:"not null"`
	MerchantName     string        `gorm:"type:varchar(50);not null"`
	CreditCardNumber uint64        `gorm:"not null"` // ForeignKey to Credit Card
	Transactions     []Transaction `gorm:"foreignKey:AuthorisationID"`
}

type Transaction struct {
	ID              uint64  `gorm:"primaryKey;autoIncrement;not null"`
	Type            string  `gorm:"type:varchar(20);not null"`
	Amount          float64 `gorm:"not null"`
	AuthorisationID string  `gorm:"not null"` // ForeignKey to Authorisation
}

type State struct {
	ID   uint64 `gorm:"primaryKey;autoIncrement;not null"`
	Name string `gorm:"type:varchar(20);not null"`
}

type Currency struct {
	ID   uint64 `gorm:"primaryKey;autoIncrement;not null"`
	Name string `gorm:"type:varchar(20);not null"`
}
