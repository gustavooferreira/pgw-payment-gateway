package core

// LuhnValid checks credit card number is valid.
func LuhnValid(creditCardNumber int64) bool {
	lastDigit := creditCardNumber % 10
	remainingDigits := creditCardNumber / 10

	var checksum int64

	for i := 0; remainingDigits > 0; i++ {
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
	checksum %= 10

	result := (lastDigit + checksum) % 10
	return result == 0
}
