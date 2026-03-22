package validator

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateStruct validates a struct based on its `binding` or `validate` tags.
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// GetValidator returns the underlying validator instance.
func GetValidator() *validator.Validate {
	return validate
}
