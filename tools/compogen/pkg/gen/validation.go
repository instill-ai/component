package gen

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

func fieldErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " field is required"
	case "len":
		return fe.Field() + " field has an invalid length"
	}

	return fe.Error() // default error
}

func asValidationError(err error) error {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return err
	}

	errs := make([]error, len(ve))
	for i, fe := range ve {
		errs[i] = fmt.Errorf(fieldErrorMessage(fe))
	}

	return errors.Join(errs...)
}
