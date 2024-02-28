package util

import (
	"github.com/opg-sirius-finance-hub/internal/sirius"
)

type pair struct {
	k string
	v string
}

var validationMappings = map[string]map[string]pair{
	"invoiceType": {
		"isEmpty": pair{"invoiceType", "Select the invoice type"},
	},
	"notes": {
		"stringLengthTooLong": pair{"notes", "The note must be 1000 characters or fewer"},
	},
}

func RenameErrors(siriusError sirius.ValidationErrors) sirius.ValidationErrors {
	mappedErrors := sirius.ValidationErrors{}
	for fieldName, value := range siriusError {
		for errorType, errorMessage := range value {
			err := make(map[string]string)
			if mapping, ok := validationMappings[fieldName][errorType]; ok {
				err[errorType] = mapping.v
				mappedErrors[mapping.k] = err
			} else {
				err[errorType] = errorMessage
				mappedErrors[fieldName] = err
			}
		}
	}
	return mappedErrors
}
