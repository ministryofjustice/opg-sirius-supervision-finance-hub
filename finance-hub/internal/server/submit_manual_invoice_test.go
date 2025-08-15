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

func TestSubmitManualInvoiceSuccess(t *testing.T) {
	form := url.Values{
		"id":                {"1"},
		"invoiceType":       {"SO"},
		"amount":            {"2025"},
		"raisedDate":        {"11/05/2024"},
		"startDate":         {"01/04/2024"},
		"endDate":           {"31/03/2025"},
		"supervisionLevel":  {"GENERAL"},
		"raised-date-day":   {""},
		"raised-date-month": {""},
		"raised-date-year":  {""},
	}

	client := mockApiClient{}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/invoices", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("clientId", "1")

	appVars := AppVars{
		Path: "/invoices",
	}

	appVars.EnvironmentVars.Prefix = "prefix"

	sut := SubmitManualInvoiceHandler{ro}

	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.Equal(t, "prefix/clients/1/invoices?success=invoice-type[SO]", w.Header().Get("HX-Redirect"))
}

func TestSubmitManualInvoiceValidationErrors(t *testing.T) {
	assert := assert.New(t)
	client := &mockApiClient{}
	ro := &mockRoute{client: client}

	validationErrors := apierror.ValidationErrors{
		"RaisedDateForAnInvoice": {
			"RaisedDateForAnInvoice": "Raised date not in the past",
		},
	}

	client.error = map[string]error{"AddManualInvoice": apierror.ValidationError{
		Errors: validationErrors,
	}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/invoices", nil)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("clientId", "1")

	appVars := AppVars{
		Path: "/add",
	}

	sut := SubmitManualInvoiceHandler{ro}
	err := sut.render(appVars, w, r)
	assert.Nil(err)
	assert.Equal("422 Unprocessable Entity", w.Result().Status)
}
