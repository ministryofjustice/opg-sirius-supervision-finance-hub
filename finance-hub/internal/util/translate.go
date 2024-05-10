package util

import "github.com/opg-sirius-finance-hub/finance-hub/internal/api"

type pair struct {
	k string
	v string
}

var validationMappings = map[string]map[string]pair{
	"adjustmentType": {
		"isEmpty": pair{"adjustmentType", "Select the adjustment type"},
	},
	"notes": {
		"isEmpty":             pair{"notes", "Enter a reason for adjustment"},
		"stringLengthTooLong": pair{"notes", "Reason for manual credit must be 1000 characters or less"},
	},
	"amount": {
		"isEmpty": pair{"amount", "Enter an amount"},
		"tooHigh": pair{"amount", "Amount entered must be less than £"},
	},
}

func RenameErrors(siriusError api.ValidationErrors) api.ValidationErrors {
	mappedErrors := api.ValidationErrors{}
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
