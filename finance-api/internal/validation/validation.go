package validation

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/opg-sirius-finance-hub/shared"
)

type Validate struct {
	validator *validator.Validate
}

func New() *Validate {
	return &Validate{
		validator: validator.New(),
	}
}

func (v *Validate) RegisterValidation(tag string, fn validator.Func) {
	err := v.validator.RegisterValidation(tag, fn)
	if err != nil {
		return
	}
}

func (v *Validate) ValidateStruct(s interface{}) shared.ValidationError {
	var validationErrors = make(shared.ValidationErrors)

	err := v.validator.Struct(s)
	if err != nil {
		errors := err.(validator.ValidationErrors)
		for _, fieldError := range errors {
			field := fieldError.Field()
			tag := fieldError.Tag()

			// Check if the map for the field exists, if not, initialize it
			if validationErrors[field] == nil {
				validationErrors[field] = make(map[string]string)
			}
			// Assign the error message to the corresponding field and tag
			validationErrors[field][tag] = fmt.Sprintf("This field %s needs to be looked at %s", field, tag)
		}

		// Construct ValidationError
		return shared.ValidationError{
			Message: "Validation error",
			Errors:  validationErrors,
		}
	}

	// No validation errors
	return shared.ValidationError{}
}

func (v *Validate) ValidateThousandCharacterCount(fl validator.FieldLevel) bool {
	return len(fl.Field().String()) < 1001
}

func (v *Validate) ValidateDateInThePast(fl validator.FieldLevel) bool {
	r := fl.Field().Interface().(shared.Date).String() // Get the string value of the field
	if r == "" {
		return false // Field is empty, consider it invalid
	}
	parsedDate, err := time.Parse("02/01/2006", r)
	if err != nil {
		return false // Error parsing date, consider it invalid
	}
	// Check if parsed date is in the past
	return parsedDate.Before(time.Now())
}
