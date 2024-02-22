package server

import (
	"github.com/opg-sirius-finance-hub/internal/model"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFeeReductions(t *testing.T) {
	feeReductions := model.FeeReductions{
		{
			Id:           1,
			Type:         "EXEMPTION",
			StartDate:    model.NewDate("2022-04-01T00:00:00+00:00"),
			EndDate:      model.NewDate("2021-03-31T00:00:00+00:00"),
			DateReceived: model.NewDate("2021-02-02T00:00:00+00:00"),
			Notes:        "Exemption cancelled due to incorrect filing",
			Deleted:      true,
		},
		{
			Id:           2,
			Type:         "REMISSION",
			StartDate:    model.NewDate("2022-04-01T00:00:00+00:00"),
			EndDate:      model.NewDate("2021-03-31T00:00:00+00:00"),
			DateReceived: model.NewDate("2021-06-02T00:00:00+00:00"),
			Notes:        "Remission for 2021/2022",
			Deleted:      false,
		},
	}

	client := mockApiClient{FeeReductions: feeReductions}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)

	appVars := AppVars{
		Path: "/1",
	}

	sut := FeeReductionsHandler{ro}

	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.True(t, ro.executed)

	expected := FeeReductionsTab{
		FeeReductions: feeReductions,
		AppVars:       appVars,
	}

	assert.Equal(t, expected, ro.data)
}
