package core

import "time"

// LuhnValid checks credit card number is valid.
func LuhnValid(creditCardNumber uint64) bool {
	var checksum uint64
	remainingDigits := creditCardNumber

	for i := 1; remainingDigits > 0; i++ {
		currentDigit := remainingDigits % 10

		if i%2 == 0 {
			currentDigit = currentDigit * 2
			if currentDigit > 9 {
				currentDigit = currentDigit%10 + currentDigit/10
			}
		}

		checksum += currentDigit
		remainingDigits /= 10
	}

	return (checksum % 10) == 0
}

func CardExpiryValid(year uint, month uint) bool {
	if month > 12 {
		return false
	}

	now := time.Now()
	nowYear := uint(now.Year())
	nowMonth := uint(now.Month())

	if year < nowYear {
		return false
	} else if year > nowYear {
		return true
	} else { // same year
		if month < nowMonth {
			return false
		} else {
			return true
		}
	}
}
