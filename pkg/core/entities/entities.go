package entities

// State should be an ENUM
type Authorisation struct {
	ID           string        `json:"id"`
	State        string        `json:"state"`
	Currency     string        `json:"currency"`
	Amount       float64       `json:"amount"`
	MerchantName string        `json:"merchant_name"`
	CreditCard   *CreditCard   `json:"credit_card,omitempty"`
	Transaction  []Transaction `json:"transactions,omitempty"`
}

type CreditCard struct {
	Number      uint64 `json:"number"`
	Name        string `json:"name"`
	ExpiryMonth uint   `json:"expiry_month"`
	ExpiryYear  uint   `json:"expiry_year"`
	CVV         uint   `json:"cvv"`
}

// Type should be an ENUM
type Transaction struct {
	ID     string  `json:"id"`
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
}
