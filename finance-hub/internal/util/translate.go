package util

import (
	"github.com/opg-sirius-finance-hub/shared"
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
		"lte":      pair{"LengthOfAward", "Award length is over 3 years"},
		"gte":      pair{"LengthOfAward", "Award length is under 1 year"},
	},
	"DateReceived": {
		"required":         pair{"DateReceived", "Enter the date received"},
		"date-in-the-past": pair{"DateReceived", "Date received must be today or in the past"},
	},
	"Notes": {
		"required":                 pair{"Notes", "Enter a reason for awarding fee reduction"},
		"thousand-character-limit": pair{"Notes", "Reason for awarding fee reduction must be 1000 characters or less"},
	},
	"Overlap": {
		"StartOrEndDate": pair{"StartOrEndDate", "A fee reduction already exists for the period specified"},
	},
}

func RenameErrors(siriusError shared.ValidationErrors) shared.ValidationErrors {
	mappedErrors := shared.ValidationErrors{}
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
