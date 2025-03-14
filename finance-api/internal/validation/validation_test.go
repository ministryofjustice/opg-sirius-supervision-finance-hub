package validation

import (
	"bytes"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"testing"
	"time"
)

type notesTest struct {
	Notes string `json:"notes" validate:"thousand-character-limit"`
}

type dateTest struct {
	DateReceived shared.Date `json:"dateReceived" validate:"date-in-the-past"`
}

type isValidEnumTest struct {
	Enum shared.AdjustmentType `json:"enum" validate:"valid-enum"`
}

func TestValidate_ValidateStruct(t *testing.T) {
	validator, _ := New()
	dateInFuture := time.Now().AddDate(0, 0, 1)
	tests := []struct {
		name        string
		args        interface{}
		expected    int
		key         string
		want        string
		description string
	}{
		{
			name:     "Count out of range of thousand",
			args:     notesTest{Notes: string(bytes.Repeat([]byte{byte('a')}, 1001))},
			expected: 1,
			key:      "Notes",
			want:     "thousand-character-limit",
		},
		{
			name:     "Count in range of thousand",
			args:     notesTest{Notes: string(bytes.Repeat([]byte{byte('a')}, 1000))},
			expected: 0,
		},
		{
			name:     "Date is not in the past",
			args:     dateTest{DateReceived: shared.Date{Time: dateInFuture}},
			expected: 1,
			key:      "DateReceived",
			want:     "date-in-the-past",
		},
		{
			name:     "Date is in the past or today",
			args:     dateTest{DateReceived: shared.Date{Time: time.Now()}},
			expected: 0,
		},
		{
			name:     "Enum type is UNKNOWN",
			args:     isValidEnumTest{Enum: shared.AdjustmentTypeUnknown},
			expected: 1,
		},
		{
			name:     "Enum type is valid",
			args:     isValidEnumTest{Enum: shared.AdjustmentTypeCreditMemo},
			expected: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.ValidateStruct(tt.args)
			if len(got.Errors) != tt.expected {
				t.Errorf("ValidateStruct() = count %v, want %v", len(got.Errors), tt.expected)
			}
			for k1, value := range got.Errors {
				if k1 == tt.key {
					for k2 := range value {
						if k2 != tt.want {
							t.Errorf("ValidateStruct() = %v, want %v", got.Errors[tt.key], tt.want)
						}
					}
				}
			}
		})
	}
}
