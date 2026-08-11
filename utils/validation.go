package utils
import (
	"github.com/go-playground/validator/v10"
	"strings"
	"errors"
)
func FormatValidationErrors(err error) []string{
	var validationError validator.ValidationErrors
	var validationmessage []string
	if errors.As(err, &validationError) {
		for _, e := range validationError {
			field := strings.ToLower(e.Field())
			tag:=e.Tag()
			switch tag {
				case "required":
					validationmessage =append(validationmessage, field + " is required")
				case "email":
					validationmessage =append(validationmessage, field + " is not a valid email")
				case "min":
					validationmessage =append(validationmessage, field + " must be at least 6 characters")
				default:
					validationmessage =append(validationmessage, field + " is invalid")
			}
		}
		return validationmessage
	}
	return nil
}