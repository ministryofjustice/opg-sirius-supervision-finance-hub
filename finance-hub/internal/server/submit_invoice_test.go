package server

import (
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSubmitInvoiceSuccess(t *testing.T) {
	form := url.Values{
		"id":          {"1"},
		"invoiceType": {"writeOff"},
		"notes":       {"This is a note"},
		"amount":      {"100"},
	}

	client := mockApiClient{}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/ledger-entries", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("id", "1")

	appVars := AppVars{
		Path: "/ledger-entries",
	}

	appVars.EnvironmentVars.Prefix = "prefix"

	sut := SubmitInvoiceHandler{ro}

	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.Equal(t, "prefix/clients/1/invoices?success=writeOff", w.Header().Get("HX-Redirect"))
}

func TestSubmitInvoiceError(t *testing.T) {
	form := url.Values{
		"id":          {"1"},
		"invoiceType": {"writeOff"},
		"notes":       {"This is a note"},
		"amount":      {"100"},
	}

	client := mockApiClient{}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/ledger-entries", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("id", "1")

	appVars := AppVars{
		Path: "/ledger-entries",
	}

	appVars.EnvironmentVars.Prefix = "prefix"

	sut := SubmitInvoiceHandler{ro}

	err := sut.render(appVars, w, r)
	assert.Nil(t, err)
	assert.Equal(t, "prefix/clients/1/invoices?success=writeOff", w.Header().Get("HX-Redirect"))
}

func TestAddTaskValidationErrors(t *testing.T) {
	assert := assert.New(t)
	client := &mockApiClient{}
	ro := &mockRoute{client: client}

	validationErrors := shared.ValidationErrors{
		"notes": {
			"stringLengthTooLong": "Reason for manual credit must be 1000 characters or less",
		},
	}

	client.error = shared.ValidationError{
		Errors: validationErrors,
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/ledger-entries", nil)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("id", "1")

	appVars := AppVars{
		Path: "/ledger-entries",
	}

	sut := SubmitInvoiceHandler{ro}
	err := sut.render(appVars, w, r)
	assert.Nil(err)
	assert.Equal("422 Unprocessable Entity", w.Result().Status)
}
