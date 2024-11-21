package util

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRenameErrors(t *testing.T) {
	siriusErrors := apierror.ValidationErrors{
		"Amount":       map[string]string{"required": ""},
		"DateReceived": map[string]string{"date-in-the-past": ""},
	}
	expected := apierror.ValidationErrors{
		"Amount":       map[string]string{"required": "Enter an amount"},
		"DateReceived": map[string]string{"date-in-the-past": "Date received must be today or in the past"},
	}

	assert.Equal(t, expected, RenameErrors(siriusErrors))
}

func TestRenameErrors_default(t *testing.T) {
	siriusErrors := apierror.ValidationErrors{
		"x": map[string]string{"y": "z"},
	}

	assert.Equal(t, siriusErrors, RenameErrors(siriusErrors))
}
