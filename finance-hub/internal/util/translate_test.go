package util

import (
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRenameErrors(t *testing.T) {
	siriusErrors := shared.ValidationErrors{
		"Amount":       map[string]string{"required": ""},
		"DateReceived": map[string]string{"date-in-the-past": ""},
	}
	expected := shared.ValidationErrors{
		"Amount":       map[string]string{"required": "Enter an amount"},
		"DateReceived": map[string]string{"date-in-the-past": "Date received must be today or in the past"},
	}

	assert.Equal(t, expected, RenameErrors(siriusErrors))
}

func TestRenameErrors_default(t *testing.T) {
	siriusErrors := shared.ValidationErrors{
		"x": map[string]string{"y": "z"},
	}

	assert.Equal(t, siriusErrors, RenameErrors(siriusErrors))
}
