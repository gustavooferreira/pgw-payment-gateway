package pprocessor

type CreditCard struct {
	Name        string `json:"name"`
	Number      uint64 `json:"number"`
	ExpiryMonth uint   `json:"expiry_month"`
	ExpiryYear  uint   `json:"expiry_year"`
	CVV         uint   `json:"cvv"`
}

type AuthorisationRequest struct {
	CreditCard CreditCard `json:"credit_card"`
	Currency   string     `json:"currency"`
	Amount     float64    `json:"amount"`
}

type AuthorisationResponse struct {
	Code            uint   `json:"code"`
	AuthorisationID string `json:"authorisation_id,omitempty"`
}

type CaptureRequest struct {
	AuthorisationID string  `json:"authorisation_id"`
	Amount          float64 `json:"amount"`
}

type CaptureResponse struct {
	Code uint `json:"code"`
}

type RefundRequest struct {
	AuthorisationID string  `json:"authorisation_id"`
	Amount          float64 `json:"amount"`
}

type RefundResponse struct {
	Code uint `json:"code"`
}

type VoidRequest struct {
	AuthorisationID string `json:"authorisation_id"`
}

type VoidResponse struct {
	Code uint `json:"code"`
}
