package util

import (
	"github.com/ministryofjustice/opg-sirius-supervision-deputy-hub/internal/sirius"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTranslate(t *testing.T) {
	type args struct {
		prefix string
		s      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Translates value when prefix and s match",
			args{"FIELD", "email"},
			"Email address",
		},
		{
			"Returns original value when no translation exists",
			args{"", "Favourite colour"},
			"Favourite colour",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Translate(tt.args.prefix, tt.args.s); got != tt.want {
				t.Errorf("Translate() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
