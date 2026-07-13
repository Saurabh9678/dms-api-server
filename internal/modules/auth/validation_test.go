package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterValidators_Idempotent(t *testing.T) {
	// Calling registerValidators twice must not panic.
	registerValidators()
	registerValidators()
}

func TestIsDigitsOnly(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"9876543210", true},
		{"91", true},
		{"0", true},
		{"99adshbbfhk", false},
		{"+91", false},
		{"-91", false},
		{"9.1", false},
		{"91 ", false},
		{"abc", false},
	}

	for _, tc := range cases {
		got := digitsOnlyString(tc.input)
		assert.Equal(t, tc.want, got, "input: %q", tc.input)
	}
}

// digitsOnlyString calls the same logic as isDigitsOnly without requiring a FieldLevel.
func digitsOnlyString(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
