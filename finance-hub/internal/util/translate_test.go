package util

import (
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRenameErrorsForEmptyValues(t *testing.T) {
	siriusErrors := shared.ValidationErrors{
		"adjustmentType": map[string]string{"isEmpty": "isEmpty"},
		"notes":          map[string]string{"isEmpty": "isEmpty"},
		"amount":         map[string]string{"isEmpty": "isEmpty"},
	}
	expected := shared.ValidationErrors{
		"adjustmentType": map[string]string{"isEmpty": "Select the adjustment type"},
		"notes":          map[string]string{"isEmpty": "Enter a reason for adjustment"},
		"amount":         map[string]string{"isEmpty": "Enter an amount"},
	}

	assert.Equal(t, expected, RenameErrors(siriusErrors))
}

func TestRenameErrorsForValuesTooLong(t *testing.T) {
	siriusErrors := shared.ValidationErrors{
		"notes": map[string]string{"stringLengthTooLong": "stringLengthTooLong"},
	}
	expected := shared.ValidationErrors{
		"notes": map[string]string{"stringLengthTooLong": "Reason for manual credit must be 1000 characters or less"},
	}

	assert.Equal(t, expected, RenameErrors(siriusErrors))
}

func TestRenameErrors_default(t *testing.T) {
	siriusErrors := shared.ValidationErrors{
		"x": map[string]string{"y": "z"},
	}

	assert.Equal(t, siriusErrors, RenameErrors(siriusErrors))
}
