package util

import (
	"github.com/opg-sirius-finance-hub/internal/sirius"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRenameErrors(t *testing.T) {
	siriusErrors := sirius.ValidationErrors{
		"name":             map[string]string{"stringLengthTooLong": "stringLengthTooLong"},
		"organisationName": map[string]string{"isEmpty": "isEmpty"},
	}
	expected := sirius.ValidationErrors{
		"1-title":          map[string]string{"stringLengthTooLong": "The title must be 255 characters or fewer"},
		"organisationName": map[string]string{"isEmpty": "Enter a deputy name"},
	}

	assert.Equal(t, expected, RenameErrors(siriusErrors))
}

func TestRenameErrors_default(t *testing.T) {
	siriusErrors := sirius.ValidationErrors{
		"x": map[string]string{"y": "z"},
	}

	assert.Equal(t, siriusErrors, RenameErrors(siriusErrors))
}
