package custom_validators

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

func AliasValidation(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case "url", "register", "login":
		return false
	default:
		return true
	}
}

func ValidationError(errs validator.ValidationErrors) error {
	var errMessages []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMessages = append(errMessages, fmt.Sprintf("field %s is required", err.Field()))
		case "url":
			errMessages = append(errMessages, fmt.Sprintf("field %s is not a valid URL", err.Field()))
		case "isValidAlias":
			errMessages = append(errMessages, "bad alias")
		default:
			errMessages = append(errMessages, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return errors.New(strings.Join(errMessages, ", "))
}
