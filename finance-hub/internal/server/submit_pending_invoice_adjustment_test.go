package server

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubmitPendingInvoiceAdjustmentSuccess(t *testing.T) {
	client := mockApiClient{}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/pending-invoice-adjustment", nil)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("ledgerId", "1")
	r.SetPathValue("clientId", "1")
	r.SetPathValue("adjustmentType", "Credit")

	appVars := AppVars{
		Path: "/pending-invoice-adjustment",
	}

	appVars.EnvironmentVars.Prefix = "prefix"

	sut := SubmitPendingInvoiceAdjustmentHandler{ro}

	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.Equal(t, "prefix/clients/1/pending-invoice-adjustments?success=credit", w.Header().Get("HX-Redirect"))
}
