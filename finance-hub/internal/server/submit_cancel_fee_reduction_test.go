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

func TestCancelFeeReductionSuccess(t *testing.T) {
	form := url.Values{
		"id":             {"1"},
		"feeReductionId": {"1"},
		"notes":          {"Fee remission note for one award"},
	}

	client := mockApiClient{}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/add", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("clientId", "1")

	appVars := AppVars{
		Path: "/cancel",
	}

	appVars.EnvironmentVars.Prefix = "prefix"

	sut := SubmitCancelFeeReductionsHandler{ro}

	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.Equal(t, "prefix/clients/1/fee-reductions?success=fee-reduction[CANCELLED]", w.Header().Get("HX-Redirect"))
}

func TestCancelFeeReductionValidationErrors(t *testing.T) {
	assert := assert.New(t)
	client := &mockApiClient{}
	ro := &mockRoute{client: client}

	validationErrors := apierror.ValidationErrors{
		"CancelFeeReductionNotes": {
			"stringLengthTooLong": "Reason for cancellation must be 1000 characters or less",
		},
	}

	client.error = map[string]error{"CancelFeeReduction": apierror.ValidationError{
		Errors: validationErrors,
	}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/add", nil)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("clientId", "1")
	r.SetPathValue("feeReductionId", "1")

	appVars := AppVars{
		Path: "/cancel",
	}

	sut := SubmitCancelFeeReductionsHandler{ro}
	err := sut.render(appVars, w, r)
	assert.Nil(err)
	assert.Equal("422 Unprocessable Entity", w.Result().Status)
}
