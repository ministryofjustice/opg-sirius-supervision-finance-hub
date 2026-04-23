package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCancelDirectDebitSuccess(t *testing.T) {
	client := mockApiClient{}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/cancel", strings.NewReader(""))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("clientId", "1")

	appVars := AppVars{
		Path: "/cancel",
	}

	appVars.EnvironmentVars.Prefix = "prefix"

	sut := SubmitCancelDirectDebitHandler{ro}

	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.Equal(t, "prefix/clients/1/invoices?success=cancel-direct-debit", w.Header().Get("HX-Redirect"))
}

func TestCancelDirectDebitErrors(t *testing.T) {
	assert := assert.New(t)
	client := &mockApiClient{}
	client.error = errors.New("this has failed")
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/cancel", strings.NewReader(""))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("clientId", "1")

	appVars := AppVars{
		Path: "/cancel",
	}

	sut := SubmitCancelDirectDebitHandler{ro}
	err := sut.render(appVars, w, r)
	assert.Nil(err)
	assert.Equal("500 Internal Server Error", w.Result().Status)
}
