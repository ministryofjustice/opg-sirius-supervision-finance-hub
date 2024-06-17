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

	validationErrors := shared.ValidationErrors{
		"RaisedDateForAnInvoice": {
			"RaisedDateForAnInvoice": "Raised date not in the past",
		},
	}

	client.error = shared.ValidationError{
		Errors: validationErrors,
	}

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

func Test_invoiceData(t *testing.T) {
	type args struct {
		invoiceType      string
		amount           string
		startDate        string
		raisedDate       string
		endDate          string
		raisedDateYear   string
		supervisionLevel string
	}
	tests := []struct {
		name             string
		args             args
		amount           string
		startDate        string
		raisedDate       string
		endDate          string
		supervisionLevel string
	}{
		{
			name: "AD invoice returns correct values",
			args: args{
				invoiceType:      "AD",
				amount:           "",
				startDate:        "",
				raisedDate:       "2023-04-01",
				endDate:          "",
				raisedDateYear:   "2023",
				supervisionLevel: "",
			},
			amount:           "100",
			startDate:        "2023-04-01",
			raisedDate:       "2023-04-01",
			endDate:          "2024-03-31",
			supervisionLevel: "",
		},
		{
			name: "S2 invoice returns correct values",
			args: args{
				invoiceType:      "S2",
				amount:           "100",
				startDate:        "2033-04-01",
				raisedDate:       "",
				endDate:          "",
				raisedDateYear:   "2025",
				supervisionLevel: "",
			},
			amount:           "100",
			startDate:        "2033-04-01",
			raisedDate:       "2025-03-31",
			endDate:          "2025-03-31",
			supervisionLevel: "GENERAL",
		},
		{
			name: "B2 invoice returns correct values",
			args: args{
				invoiceType:      "B2",
				amount:           "100",
				startDate:        "2033-04-01",
				raisedDate:       "",
				endDate:          "",
				raisedDateYear:   "2025",
				supervisionLevel: "",
			},
			amount:           "100",
			startDate:        "2033-04-01",
			raisedDate:       "2025-03-31",
			endDate:          "2025-03-31",
			supervisionLevel: "GENERAL",
		},
		{
			name: "S3 invoice returns correct values",
			args: args{
				invoiceType:      "S3",
				amount:           "100",
				startDate:        "2033-04-01",
				raisedDate:       "",
				endDate:          "",
				raisedDateYear:   "2025",
				supervisionLevel: "",
			},
			amount:           "100",
			startDate:        "2033-04-01",
			raisedDate:       "2025-03-31",
			endDate:          "2025-03-31",
			supervisionLevel: "MINIMAL",
		},
		{
			name: "B3 invoice returns correct values",
			args: args{
				invoiceType:      "S3",
				amount:           "100",
				startDate:        "2033-04-01",
				raisedDate:       "",
				endDate:          "",
				raisedDateYear:   "2025",
				supervisionLevel: "",
			},
			amount:           "100",
			startDate:        "2033-04-01",
			raisedDate:       "2025-03-31",
			endDate:          "2025-03-31",
			supervisionLevel: "MINIMAL",
		},
		{
			name: "No year will return correct values",
			args: args{
				invoiceType:      "S3",
				amount:           "100",
				startDate:        "2033-04-01",
				raisedDate:       "",
				endDate:          "",
				raisedDateYear:   "",
				supervisionLevel: "",
			},
			amount:           "100",
			startDate:        "2033-04-01",
			raisedDate:       "",
			endDate:          "",
			supervisionLevel: "MINIMAL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, got3, got4 := invoiceData(tt.args.invoiceType, tt.args.amount, tt.args.startDate, tt.args.raisedDate, tt.args.endDate, tt.args.raisedDateYear, tt.args.supervisionLevel)
			assert.Equalf(t, tt.amount, got, "invoiceData(%v, %v, %v, %v, %v, %v, %v)", tt.args.invoiceType, tt.args.amount, tt.args.startDate, tt.args.raisedDate, tt.args.endDate, tt.args.raisedDateYear, tt.args.supervisionLevel)
			assert.Equalf(t, tt.startDate, got1, "invoiceData(%v, %v, %v, %v, %v, %v, %v)", tt.args.invoiceType, tt.args.amount, tt.args.startDate, tt.args.raisedDate, tt.args.endDate, tt.args.raisedDateYear, tt.args.supervisionLevel)
			assert.Equalf(t, tt.raisedDate, got2, "invoiceData(%v, %v, %v, %v, %v, %v, %v)", tt.args.invoiceType, tt.args.amount, tt.args.startDate, tt.args.raisedDate, tt.args.endDate, tt.args.raisedDateYear, tt.args.supervisionLevel)
			assert.Equalf(t, tt.endDate, got3, "invoiceData(%v, %v, %v, %v, %v, %v, %v)", tt.args.invoiceType, tt.args.amount, tt.args.startDate, tt.args.raisedDate, tt.args.endDate, tt.args.raisedDateYear, tt.args.supervisionLevel)
			assert.Equalf(t, tt.supervisionLevel, got4, "invoiceData(%v, %v, %v, %v, %v, %v, %v)", tt.args.invoiceType, tt.args.amount, tt.args.startDate, tt.args.raisedDate, tt.args.endDate, tt.args.raisedDateYear, tt.args.supervisionLevel)
		})
	}
}
