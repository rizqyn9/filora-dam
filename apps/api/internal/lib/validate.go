package lib

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

// Validate runs struct validation and returns a client-friendly AppError on
// failure, listing the offending fields.
func Validate(s any) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var invalid validator.ValidationErrors
	if !asValidationErrors(err, &invalid) {
		return ErrValidation("invalid input").Wrap(err)
	}

	fields := make([]string, 0, len(invalid))
	for _, fe := range invalid {
		fields = append(fields, fe.Field())
	}
	return ErrValidation("validation failed for: " + strings.Join(fields, ", ")).Wrap(err)
}

func asValidationErrors(err error, target *validator.ValidationErrors) bool {
	if ve, ok := err.(validator.ValidationErrors); ok {
		*target = ve
		return true
	}
	return false
}
