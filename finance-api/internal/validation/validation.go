package validation

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"strconv"
	"strings"
	"time"
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
	err = v.RegisterValidation("nillable-int-gt", validateIntGreaterThan)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("nillable-int-lte", validateIntLessThanOrEqualTo)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("nillable-date-required", validateDateRequiredIfNotNil)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("nillable-string-oneof", validateStringOneOf)
	if err != nil {
		return nil, err
	}
	return &Validate{
		validator: v,
	}, nil
}

func (v *Validate) ValidateStruct(s interface{}) apierror.ValidationError {
	var validationErrors = make(apierror.ValidationErrors)

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

		return apierror.ValidationError{
			Errors: validationErrors,
		}
	}

	return apierror.ValidationError{}
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
	if v, ok := fl.Field().Interface().(shared.Enum); ok {
		if v.Valid() {
			return true
		}
	}
	return false
}

func validateIntGreaterThan(fl validator.FieldLevel) bool {
	if v, ok := fl.Field().Interface().(shared.Nillable[int]); ok {
		intParam, err := strconv.Atoi(fl.Param())
		if err != nil {
			panic(err)
		}
		return !v.Valid || v.Value > intParam
	} else if v, ok := fl.Field().Interface().(shared.Nillable[int32]); ok {
		intParam, err := strconv.ParseInt(fl.Param(), 10, 32)
		if err != nil {
			panic(err)
		}
		return !v.Valid || v.Value > int32(intParam)
	}
	return false
}

func validateIntLessThanOrEqualTo(fl validator.FieldLevel) bool {
	if v, ok := fl.Field().Interface().(shared.Nillable[int]); ok {
		intParam, err := strconv.Atoi(fl.Param())
		if err != nil {
			panic(err)
		}
		return !v.Valid || v.Value <= intParam
	} else if v, ok := fl.Field().Interface().(shared.Nillable[int32]); ok {
		intParam, err := strconv.ParseInt(fl.Param(), 10, 32)
		if err != nil {
			panic(err)
		}
		return !v.Valid || v.Value <= int32(intParam)
	}
	return false
}

func validateStringOneOf(fl validator.FieldLevel) bool {
	if v, ok := fl.Field().Interface().(shared.Nillable[string]); ok {
		if !v.Valid {
			return true
		}
		params := strings.Fields(fl.Param())
		for _, param := range params {
			if param == v.Value {
				return true
			}
		}
	}
	return false
}

func validateDateRequiredIfNotNil(fl validator.FieldLevel) bool {
	if v, ok := fl.Field().Interface().(shared.Nillable[shared.Date]); ok {
		if !v.Valid || !v.Value.IsNull() {
			return true
		}
	}
	return false
}
