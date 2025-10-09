package server

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDirectDebitMandate(t *testing.T) {
	client := mockApiClient{}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("clientId", "1")

	appVars := AppVars{Path: "/path/"}

	sut := DirectDebitMandateHandler{ro}
	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.True(t, ro.executed)

	expected := DirectDebitMandateForm{
		"1",
		appVars,
	}
	assert.Equal(t, expected, ro.data)
}
