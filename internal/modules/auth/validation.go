package auth

import (
	"sync"

	"github.com/gin-gonic/gin/binding"
	validator "github.com/go-playground/validator/v10"
)

var registerValidatorsOnce sync.Once

func init() {
	registerValidators()
}

func registerValidators() {
	registerValidatorsOnce.Do(func() {
		v, ok := binding.Validator.Engine().(*validator.Validate)
		if !ok {
			return
		}
		_ = v.RegisterValidation("digits", isDigitsOnly)
	})
}

// isDigitsOnly accepts strings that contain only ASCII digit characters (0–9).
// It rejects sign prefixes (+/-), decimal points, spaces, and any non-digit character.
func isDigitsOnly(fl validator.FieldLevel) bool {
	for _, c := range fl.Field().String() {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
