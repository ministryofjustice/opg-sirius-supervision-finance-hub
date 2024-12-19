package server

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAddDirectDebitSuccess(t *testing.T) {
	form := url.Values{
		"accountHolder": {"CLIENT"},
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

	sut := SubmitDirectDebitHandler{ro}

	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.Equal(t, "prefix/clients/1/invoices?success=direct-debit", w.Header().Get("HX-Redirect"))
}

func TestDirectDebit_Errors(t *testing.T) {
	client := mockApiClient{}
	client.error = errors.New("this has failed")
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("clientId", "1")

	appVars := AppVars{Path: "/path/"}

	sut := SubmitFeeReductionsHandler{ro}
	err := sut.render(appVars, w, r)

	assert.Equal(t, "this has failed", err.Error())
	assert.False(t, ro.executed)
}
