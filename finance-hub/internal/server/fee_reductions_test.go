package server

import (
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFeeReductions(t *testing.T) {
	in := shared.FeeReductions{
		{
			Id:                       1,
			Type:                     "EXEMPTION",
			StartDate:                shared.NewDate("2022-04-01T00:00:00+00:00"),
			EndDate:                  shared.NewDate("2021-03-31T00:00:00+00:00"),
			DateReceived:             shared.NewDate("2021-02-02T00:00:00+00:00"),
			Status:                   "Expired",
			Notes:                    "Exemption cancelled due to incorrect filing",
			FeeReductionCancelAction: false,
		},
		{
			Id:                       2,
			Type:                     "REMISSION",
			StartDate:                shared.NewDate("2022-04-01T00:00:00+00:00"),
			EndDate:                  shared.NewDate("2021-03-31T00:00:00+00:00"),
			DateReceived:             shared.NewDate("2021-06-02T00:00:00+00:00"),
			Status:                   "Active",
			Notes:                    "Remission for 2021/2022",
			FeeReductionCancelAction: true,
		},
	}

	client := mockApiClient{FeeReductions: in}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("id", "1")

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
			ClientId:                 "1",
		},
		{
			Type:                     "Remission",
			StartDate:                "01/04/2022",
			EndDate:                  "31/03/2021",
			DateReceived:             "02/06/2021",
			Status:                   "Active",
			Notes:                    "Remission for 2021/2022",
			FeeReductionCancelAction: true,
			Id:                       "2",
			ClientId:                 "1",
		},
	}

	expected := &FeeReductionsTab{
		FeeReductions: out,
		ClientId:      "1",
		AppVars:       appVars,
	}

	assert.Equal(t, expected, ro.data)
}
