package server

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAddFeeReductionSuccess(t *testing.T) {
	form := url.Values{
		"id":            {"1"},
		"feeType":       {"remission"},
		"startYear":     {"2025"},
		"lengthOfAward": {"3"},
		"dateReceived":  {"15/02/2024"},
		"notes":         {"Fee remission note for one award"},
	}

	client := mockApiClient{}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/add", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("clientId", "1")

	appVars := AppVars{
		Path: "/add",
	}

	appVars.EnvironmentVars.Prefix = "prefix"

	sut := SubmitFeeReductionsHandler{ro}

	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.Equal(t, "prefix/clients/1/fee-reductions?success=fee-reduction[REMISSION]", w.Header().Get("HX-Redirect"))
}

func TestAddFeeReductionValidationErrors(t *testing.T) {
	assert := assert.New(t)
	client := &mockApiClient{}
	ro := &mockRoute{client: client}

	validationErrors := apierror.ValidationErrors{
		"notes": {
			"stringLengthTooLong": "Reason for manual credit must be 1000 characters or less",
		},
	}

	client.error = map[string]error{"AddFeeReduction": apierror.ValidationError{
		Errors: validationErrors,
	}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/add", nil)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("clientId", "1")

	appVars := AppVars{
		Path: "/add",
	}

	sut := SubmitFeeReductionsHandler{ro}
	err := sut.render(appVars, w, r)
	assert.Nil(err)
	assert.Equal("422 Unprocessable Entity", w.Result().Status)
}
