package shared

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

// test to prove unmarshalling of empty interfaces to types based on value
func TestBillingHistory_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		in   BillingHistory
		err  bool
	}{
		{
			name: "valid invoice generated",
			in: BillingHistory{
				Event: &InvoiceGenerated{
					BaseBillingEvent: BaseBillingEvent{
						Type: EventTypeInvoiceGenerated,
					},
					InvoiceReference: InvoiceReference{
						ID:        123,
						Reference: "123abc",
					},
					InvoiceType: "SE",
					Amount:      1001,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := json.Marshal(tt.in)
			var be BillingHistory
			err := be.UnmarshalJSON(data)
			assert.Equal(t, tt.err, err != nil)
			if err == nil {
				assert.EqualValues(t, tt.in, be)
			}
		})
	}
}
