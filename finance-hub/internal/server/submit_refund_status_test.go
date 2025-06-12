package server

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSubmitRefundSuccess(t *testing.T) {
	client := mockApiClient{}
	ro := &mockRoute{client: client}

	form := url.Values{
		"status": {"APPROVED"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPut, "/clients/1/refunds/2", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("clientId", "1")
	r.SetPathValue("refundId", "2")

	appVars := AppVars{
		Path: "/clients/1/refunds",
	}

	appVars.EnvironmentVars.Prefix = "prefix"

	sut := SubmitRefundStatusHandler{ro}

	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.Equal(t, "prefix/clients/1/refunds?success=refunds[APPROVED]", w.Header().Get("HX-Redirect"))
}
