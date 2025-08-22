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

func TestSetupDirectDebitSuccess(t *testing.T) {
	form := url.Values{
		"accountName":   {"account name"},
		"sortCode":      {"123456"},
		"accountNumber": {"12345678"},
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

	sut := SetupDirectDebitHandler{ro}

	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.Equal(t, "prefix/clients/1/invoices?success=direct-debit", w.Header().Get("HX-Redirect"))
}

func TestSetupDirectDebitValidationErrors(t *testing.T) {
	assert := assert.New(t)
	client := &mockApiClient{}
	ro := &mockRoute{client: client}

	validationErrors := apierror.ValidationErrors{
		"SortCode": {
			"required": "Sort code must be provided",
		},
	}

	client.error = map[string]error{"CreateDirectDebitMandate": apierror.ValidationError{
		Errors: validationErrors,
	}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/add", nil)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("clientId", "1")

	appVars := AppVars{
		Path: "/add",
	}

	sut := SetupDirectDebitHandler{ro}
	err := sut.render(appVars, w, r)
	assert.Nil(err)
	assert.Equal("422 Unprocessable Entity", w.Result().Status)
}

func TestSetupDirect_failToCreateSchedule(t *testing.T) {
	assert := assert.New(t)
	client := &mockApiClient{}
	ro := &mockRoute{client: client}

	validationErrors := apierror.ValidationErrors{
		"SortCode": {
			"required": "Sort code must be provided",
		},
	}

	client.error = map[string]error{"CreateDirectDebitSchedule": apierror.ValidationError{
		Errors: validationErrors,
	}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/add", nil)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("clientId", "1")

	appVars := AppVars{
		Path: "/add",
	}

	sut := SetupDirectDebitHandler{ro}
	err := sut.render(appVars, w, r)
	assert.Nil(err)
	assert.Equal("422 Unprocessable Entity", w.Result().Status)
}
