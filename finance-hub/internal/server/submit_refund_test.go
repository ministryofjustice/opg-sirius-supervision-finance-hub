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

func TestAddRefundSuccess(t *testing.T) {
	form := url.Values{
		"id":            {"1"},
		"accountName":   {"Mr Reginald Refund"},
		"accountNumber": {"12345678"},
		"sortCode":      {"11-22-33"},
		"Notes":         {"notes here"},
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

	sut := SubmitRefundHandler{ro}

	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.Equal(t, "prefix/clients/1/refunds?success=refund", w.Header().Get("HX-Redirect"))
}

func TestAddRefundValidationErrors(t *testing.T) {
	assert := assert.New(t)
	client := &mockApiClient{}
	ro := &mockRoute{client: client}

	validationErrors := apierror.ValidationErrors{
		"accountNumber": {
			"tooLong": "Account number must by 8 digits",
		},
	}

	client.error = apierror.ValidationError{
		Errors: validationErrors,
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/add", nil)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("clientId", "1")

	appVars := AppVars{
		Path: "/add",
	}

	sut := SubmitRefundHandler{ro}
	err := sut.render(appVars, w, r)
	assert.Nil(err)
	assert.Equal("422 Unprocessable Entity", w.Result().Status)
}
