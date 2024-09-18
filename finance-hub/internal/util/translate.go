package util

import (
	"github.com/opg-sirius-finance-hub/apierror"
)

type pair struct {
	k string
	v string
}

var validationMappings = map[string]map[string]pair{
	"AdjustmentType": {
		"required":   pair{"AdjustmentType", "Select the adjustment type"},
		"valid-enum": pair{"AdjustmentType", "Select the adjustment type"},
	},
	"AdjustmentNotes": {
		"required":                 pair{"AdjustmentNotes", "Enter a reason for adjustment"},
		"stringLengthTooLong":      pair{"AdjustmentNotes", "Reason for manual credit must be 1000 characters or less"},
		"thousand-character-limit": pair{"Notes", "Reason for adjustment must be 1000 characters or less"},
	},
	"Amount": {
		"required":         pair{"Amount", "Enter an amount"},
		"required_if":      pair{"Amount", "Enter an amount"},
		"nillable-int-lte": pair{"Amount", "Amount can't be above £320"},
		"nillable-int-gt":  pair{"Amount", "Enter an amount"},
	},
	"FeeType": {
		"required": pair{"FeeType", "A fee reduction type must be selected"},
	},
	"StartYear": {
		"required": pair{"StartYear", "Enter a start year"},
	},
	"RaisedDate": {
		"nillable-date-required": pair{"RaisedDate", "Enter a raised date"},
		"date-in-the-past":       pair{"RaisedDate", "Enter a raised date in the past"},
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
	"CancellationReason": {
		"required":                 pair{"CancellationReason", "Enter a reason for cancelling fee reduction"},
		"thousand-character-limit": pair{"CancellationReason", "Reason for cancellation must be 1000 characters or less"},
	},
	"Overlap": {
		"start-or-end-date": pair{"start-or-end-date", "A fee reduction already exists for the period specified"},
	},
	"StartDate": {
		"StartDate":              pair{"StartDate", "Start date must be before end date and in same financial year"},
		"nillable-date-required": pair{"StartDate", "Enter a start date"},
	},
	"EndDate": {
		"EndDate":                pair{"EndDate", "End date must be after start date and in same financial year"},
		"nillable-date-required": pair{"EndDate", "Enter an end date"},
	},
	"RaisedDateNotInPast": {
		"RaisedDateNotInPast": pair{"RaisedDateNotInPast", "Raised date not in the past"},
	},
	"RaisedDateForAnInvoice": {
		"RaisedDateForAnInvoice": pair{"RaisedDateForAnInvoice", "Raised date not in the past"},
	},
	"InvoiceType": {
		"required": pair{"InvoiceType", "Please select an invoice type"},
	},
	"SupervisionLevel": {
		"nillable-string-oneof": pair{"SupervisionLevel", "Please select a valid supervision level"},
	},
}

func RenameErrors(siriusError apierror.ValidationErrors) apierror.ValidationErrors {
	mappedErrors := apierror.ValidationErrors{}
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
