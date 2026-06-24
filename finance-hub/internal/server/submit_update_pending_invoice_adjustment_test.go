package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubmitApprovePendingInvoiceAdjustmentSuccess(t *testing.T) {
	client := mockApiClient{}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/pending-invoice-adjustment", strings.NewReader(""))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("ledgerId", "1")
	r.SetPathValue("clientId", "1")
	r.SetPathValue("status", "approved")
	r.SetPathValue("adjustmentType", "Credit")

	appVars := AppVars{
		Path: "/pending-invoice-adjustment",
	}

	appVars.EnvironmentVars.Prefix = "prefix"

	sut := SubmitUpdatePendingInvoiceAdjustmentHandler{ro}

	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.Equal(t, "prefix/clients/1/invoice-adjustments?success=approved-invoice-adjustment[CREDIT]", w.Header().Get("HX-Redirect"))
}

func TestSubmitRejectPendingInvoiceAdjustmentSuccess(t *testing.T) {
	client := mockApiClient{}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/pending-invoice-adjustment", strings.NewReader(""))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("ledgerId", "1")
	r.SetPathValue("clientId", "1")
	r.SetPathValue("status", "rejected")
	r.SetPathValue("adjustmentType", "Credit")

	appVars := AppVars{
		Path: "/pending-invoice-adjustment",
	}

	appVars.EnvironmentVars.Prefix = "prefix"

	sut := SubmitUpdatePendingInvoiceAdjustmentHandler{ro}

	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.Equal(t, "prefix/clients/1/invoice-adjustments?success=rejected-invoice-adjustment[CREDIT]", w.Header().Get("HX-Redirect"))
}
