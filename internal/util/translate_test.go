package util

import (
	"github.com/opg-sirius-finance-hub/internal/sirius"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRenameErrors(t *testing.T) {
	siriusErrors := sirius.ValidationErrors{
		"invoiceType": map[string]string{"isEmpty": "isEmpty"},
		"notes":       map[string]string{"stringLengthTooLong": "stringLengthTooLong"},
	}
	expected := sirius.ValidationErrors{
		"invoiceType": map[string]string{"isEmpty": "Select the invoice type"},
		"notes":       map[string]string{"stringLengthTooLong": "The note must be 1000 characters or fewer"},
	}

	assert.Equal(t, expected, RenameErrors(siriusErrors))
}

func TestRenameErrors_default(t *testing.T) {
	siriusErrors := sirius.ValidationErrors{
		"x": map[string]string{"y": "z"},
	}

	assert.Equal(t, siriusErrors, RenameErrors(siriusErrors))
}
