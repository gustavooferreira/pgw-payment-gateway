package core

// LuhnValid checks credit card number is valid.
func LuhnValid(creditCardNumber int64) bool {
	var checksum int64
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
