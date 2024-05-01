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
	"FeeType": {
		"required": pair{"FeeType", "A fee reduction type must be selected"},
	},
	"StartYear": {
		"required": pair{"StartYear", "Enter a start year"},
	},
	"LengthOfAward": {
		"required": pair{"LengthOfAward", "Confirm if an extended award is being given"},
	},
	"DateReceive": {
		"required":          pair{"DateReceive", "Enter the date received"},
		"check-in-the-past": pair{"DateReceive", "Date received must be in the past"},
	},
	"FeeReductionNotes": {
		"required":              pair{"FeeReductionNotes", "Enter a reason for awarding fee reduction"},
		"check-character-limit": pair{"FeeReductionNotes", "Reason for awarding fee reduction must be 1000 characters or less"},
	},
	"Overlap": {
		"StartOrEndDate": pair{"StartOrEndDate", "A fee reduction already exists for the period specified"},
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
