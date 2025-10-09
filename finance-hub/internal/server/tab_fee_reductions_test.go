package server

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFeeReductions(t *testing.T) {
	in := shared.FeeReductions{
		{
			Id:           1,
			Type:         shared.FeeReductionTypeExemption,
			StartDate:    shared.NewDate("2022-04-01T00:00:00+00:00"),
			EndDate:      shared.NewDate("2021-03-31T00:00:00+00:00"),
			DateReceived: shared.NewDate("2021-02-02T00:00:00+00:00"),
			Status:       "Expired",
			Notes:        "Exemption cancelled due to incorrect filing",
		},
		{
			Id:           2,
			Type:         shared.FeeReductionTypeRemission,
			StartDate:    shared.NewDate("2022-04-01T00:00:00+00:00"),
			EndDate:      shared.NewDate("2021-03-31T00:00:00+00:00"),
			DateReceived: shared.NewDate("2021-06-02T00:00:00+00:00"),
			Status:       shared.StatusActive,
			Notes:        "Remission for 2021/2022",
		},
	}

	client := mockApiClient{FeeReductions: in}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("clientId", "1")

	appVars := AppVars{
		Path: "/1",
	}

	sut := FeeReductionsHandler{ro}

	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.True(t, ro.executed)

	out := FeeReductions{
		{
			Type:                     "Exemption",
			StartDate:                "01/04/2022",
			EndDate:                  "31/03/2021",
			DateReceived:             "02/02/2021",
			Status:                   "Expired",
			Notes:                    "Exemption cancelled due to incorrect filing",
			FeeReductionCancelAction: false,
			Id:                       "1",
		},
		{
			Type:                     "Remission",
			StartDate:                "01/04/2022",
			EndDate:                  "31/03/2021",
			DateReceived:             "02/06/2021",
			Status:                   shared.StatusActive,
			Notes:                    "Remission for 2021/2022",
			FeeReductionCancelAction: true,
			Id:                       "2",
		},
	}

	expected := &FeeReductionsTab{
		FeeReductions: out,
		ClientId:      "1",
		AppVars:       appVars,
	}

	assert.Equal(t, expected, ro.data)
}

func Test_showFeeReductionCancelBtn(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "returns false for expired status",
			status: "Expired",
			want:   false,
		},
		{
			name:   "returns false for cancelled status",
			status: "Cancelled",
			want:   false,
		},
		{
			name:   "returns true for active status",
			status: shared.StatusActive,
			want:   true,
		},
		{
			name:   "returns true for pending status",
			status: shared.StatusPending,
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, showFeeReductionCancelBtn(tt.status), "showFeeReductionCancelBtn(%v)", tt.status)
		})
	}
}
