package util

import (
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRenameErrorsForEmptyValues(t *testing.T) {
	siriusErrors := shared.ValidationErrors{
		"invoiceType": map[string]string{"isEmpty": "isEmpty"},
		"notes":       map[string]string{"isEmpty": "isEmpty"},
		"amount":      map[string]string{"isEmpty": "isEmpty"},
	}
	expected := shared.ValidationErrors{
		"invoiceType": map[string]string{"isEmpty": "Select the invoice type"},
		"notes":       map[string]string{"isEmpty": "Enter a reason for adjustment"},
		"amount":      map[string]string{"isEmpty": "Enter an amount"},
	}

	assert.Equal(t, expected, RenameErrors(siriusErrors))
}

func TestRenameErrorsForValuesTooHigh(t *testing.T) {
	siriusErrors := shared.ValidationErrors{
		"notes":  map[string]string{"stringLengthTooLong": "stringLengthTooLong"},
		"amount": map[string]string{"tooHigh": "Amount entered must be less than £"},
	}
	expected := shared.ValidationErrors{
		"notes":  map[string]string{"stringLengthTooLong": "Reason for manual credit must be 1000 characters or less"},
		"amount": map[string]string{"tooHigh": "Amount entered must be less than £"},
	}

	assert.Equal(t, expected, RenameErrors(siriusErrors))
}

func TestRenameErrors_default(t *testing.T) {
	siriusErrors := shared.ValidationErrors{
		"x": map[string]string{"y": "z"},
	}

	assert.Equal(t, siriusErrors, RenameErrors(siriusErrors))
}
