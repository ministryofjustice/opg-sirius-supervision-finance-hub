package util

import (
	"github.com/opg-sirius-finance-hub/finance-hub/internal/api"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRenameErrorsForEmptyValues(t *testing.T) {
	siriusErrors := api.ValidationErrors{
		"invoiceType": map[string]string{"isEmpty": "isEmpty"},
		"notes":       map[string]string{"isEmpty": "isEmpty"},
		"amount":      map[string]string{"isEmpty": "isEmpty"},
	}
	expected := api.ValidationErrors{
		"invoiceType": map[string]string{"isEmpty": "Select the invoice type"},
		"notes":       map[string]string{"isEmpty": "Enter a reason for adjustment"},
		"amount":      map[string]string{"isEmpty": "Enter an amount"},
	}

	assert.Equal(t, expected, RenameErrors(siriusErrors))
}

func TestRenameErrorsForValuesTooHigh(t *testing.T) {
	siriusErrors := api.ValidationErrors{
		"notes":  map[string]string{"stringLengthTooLong": "stringLengthTooLong"},
		"amount": map[string]string{"tooHigh": "Amount entered must be less than £"},
	}
	expected := api.ValidationErrors{
		"notes":  map[string]string{"stringLengthTooLong": "Reason for manual credit must be 1000 characters or less"},
		"amount": map[string]string{"tooHigh": "Amount entered must be less than £"},
	}

	assert.Equal(t, expected, RenameErrors(siriusErrors))
}

func TestRenameErrors_default(t *testing.T) {
	siriusErrors := api.ValidationErrors{
		"x": map[string]string{"y": "z"},
	}

	assert.Equal(t, siriusErrors, RenameErrors(siriusErrors))
}
