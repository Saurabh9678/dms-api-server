package auth

import (
	"regexp"
	"sync"

	"github.com/gin-gonic/gin/binding"
	validator "github.com/go-playground/validator/v10"
)

var registerValidatorsOnce sync.Once
var digitsOnlyPattern = regexp.MustCompile(`^[0-9]+$`)

func init() {
	registerValidators()
}

func registerValidators() {
	registerValidatorsOnce.Do(func() {
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			_ = v.RegisterValidation("digits", isDigitsOnly)
		}
	})
}

// isDigitsOnly accepts strings that contain only ASCII digit characters (0-9).
// It rejects sign prefixes (+/-), decimal points, spaces, and any non-digit character.
func isDigitsOnly(fl validator.FieldLevel) bool {
	return digitsOnlyPattern.MatchString(fl.Field().String())
}
