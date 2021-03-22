package core_test

import (
	"testing"

	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestLuhnValid(t *testing.T) {
	tests := map[string]struct {
		creditCardNumber int64
		expectedOutput   bool
	}{
		"valid credit card 1":   {creditCardNumber: 4000000000000119, expectedOutput: true},
		"valid credit card 2":   {creditCardNumber: 4000000000000259, expectedOutput: true},
		"valid credit card 3":   {creditCardNumber: 4000000000003238, expectedOutput: true},
		"invalid credit card 1": {creditCardNumber: 4000000000000009, expectedOutput: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			value := core.LuhnValid(test.creditCardNumber)
			assert.Equal(t, test.expectedOutput, value)
		})
	}
}
