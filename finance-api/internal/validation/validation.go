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

func New() (*Validate, error) {
	v := validator.New(validator.WithRequiredStructEnabled())
	err := v.RegisterValidation("thousand-character-limit", validateThousandCharacterCount)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("date-in-the-past", validateDateInThePast)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("valid-enum", validateEnum)
	if err != nil {
		return nil, err
	}
	return &Validate{
		validator: v,
	}, nil
}

func (v *Validate) ValidateStruct(s interface{}) shared.ValidationError {
	var validationErrors = make(shared.ValidationErrors)

	err := v.validator.Struct(s)
	if err != nil {
		errors := err.(validator.ValidationErrors)
		for _, fieldError := range errors {
			field := fieldError.Field()
			tag := fieldError.Tag()

			if validationErrors[field] == nil {
				validationErrors[field] = make(map[string]string)
			}
			validationErrors[field][tag] = fmt.Sprintf("This field %s needs to be looked at %s", field, tag)
		}

		return shared.ValidationError{
			Errors: validationErrors,
		}
	}

	return shared.ValidationError{}
}

func validateThousandCharacterCount(fl validator.FieldLevel) bool {
	return len(fl.Field().String()) < 1001
}

func validateDateInThePast(fl validator.FieldLevel) bool {
	d := fl.Field().Interface().(shared.Date)
	r := d.String()
	if r == "" {
		return false // Field is empty, consider it invalid
	}

	parsedDate, err := time.Parse("02/01/2006", r)
	if err != nil {
		return false // Error parsing date, consider it invalid
	}

	return parsedDate.Before(time.Now())
}

func validateEnum(fl validator.FieldLevel) bool {
	if v, ok := fl.Field().Interface().(shared.Valid); ok {
		if v.Valid() {
			return true
		}
	}
	return false
}
