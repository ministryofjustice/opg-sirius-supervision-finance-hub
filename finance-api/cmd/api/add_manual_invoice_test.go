package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/opg-sirius-finance-hub/finance-api/internal/service"
	"github.com/opg-sirius-finance-hub/finance-api/internal/validation"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServer_addManualInvoice(t *testing.T) {
	var b bytes.Buffer
	var startDateTransformed *shared.Date
	var endDateTransformed *shared.Date

	startDateToTime, _ := time.Parse("2006-01-02", "2024-04-12")
	startDateTransformed = &shared.Date{Time: startDateToTime}

	endDateToTime, _ := time.Parse("2006-01-02", "2025-03-31")
	endDateTransformed = &shared.Date{Time: endDateToTime}

	manualInvoiceInfo := &shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeS2,
		Amount:           32000,
		RaisedDate:       endDateTransformed,
		StartDate:        startDateTransformed,
		EndDate:          endDateTransformed,
		SupervisionLevel: "GENERAL",
	}
	_ = json.NewEncoder(&b).Encode(manualInvoiceInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/invoices", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{manualInvoice: manualInvoiceInfo}
	server := Server{Service: mock, Validator: validator}
	server.addManualInvoice(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	expected := ""

	assert.Equal(t, expected, string(data))
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestServer_addManualInvoiceValidationErrors(t *testing.T) {
	var b bytes.Buffer
	manualInvoiceInfo := &shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeUnknown,
		Amount:           0,
		RaisedDate:       nil,
		StartDate:        nil,
		EndDate:          nil,
		SupervisionLevel: "",
	}
	_ = json.NewEncoder(&b).Encode(manualInvoiceInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/invoices", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{manualInvoice: manualInvoiceInfo}
	server := Server{Service: mock, Validator: validator}
	server.addManualInvoice(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	expected := `
{"Message":"","validation_errors":{"Amount":{"required":"This field Amount needs to be looked at required"},"EndDate":{"required":"This field EndDate needs to be looked at required"},"InvoiceType":{"required":"This field InvoiceType needs to be looked at required"},"RaisedDate":{"required":"This field RaisedDate needs to be looked at required"},"StartDate":{"required":"This field StartDate needs to be looked at required"}}}`

	assert.Equal(t, expected, string(data))
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestServer_addManualInvoiceValidationErrorsForAmountTooHigh(t *testing.T) {
	var b bytes.Buffer
	var startDateTransformed *shared.Date
	var endDateTransformed *shared.Date

	startDateToTime, _ := time.Parse("2006-01-02", "2024-04-12")
	startDateTransformed = &shared.Date{Time: startDateToTime}

	endDateToTime, _ := time.Parse("2006-01-02", "2025-03-31")
	endDateTransformed = &shared.Date{Time: endDateToTime}
	manualInvoiceInfo := &shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeS2,
		Amount:           320000,
		RaisedDate:       endDateTransformed,
		StartDate:        startDateTransformed,
		EndDate:          endDateTransformed,
		SupervisionLevel: "GENERAL",
	}
	_ = json.NewEncoder(&b).Encode(manualInvoiceInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/invoices", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{manualInvoice: manualInvoiceInfo}
	server := Server{Service: mock, Validator: validator}
	server.addManualInvoice(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	expected := `
{"Message":"","validation_errors":{"Amount":{"lte":"This field Amount needs to be looked at lte"}}}`

	assert.Equal(t, expected, string(data))
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestServer_addManualInvoiceDateErrors(t *testing.T) {
	var b bytes.Buffer
	var startDateTransformed *shared.Date
	var endDateTransformed *shared.Date

	startDateToTime, _ := time.Parse("2006-01-02", "2024-04-12")
	startDateTransformed = &shared.Date{Time: startDateToTime}

	endDateToTime, _ := time.Parse("2006-01-02", "2025-03-31")
	endDateTransformed = &shared.Date{Time: endDateToTime}
	manualInvoiceInfo := &shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeS2,
		Amount:           320000,
		RaisedDate:       endDateTransformed,
		StartDate:        startDateTransformed,
		EndDate:          endDateTransformed,
		SupervisionLevel: "GENERAL",
	}
	_ = json.NewEncoder(&b).Encode(manualInvoiceInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/invoices", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{manualInvoice: manualInvoiceInfo, err: service.BadRequest{Reason: " RaisedDateForAnInvoice, StartDate, EndDate"}}
	server := Server{Service: mock, Validator: validator}
	server.addFeeReduction(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestServer_addManualInvoice422Error(t *testing.T) {
	var b bytes.Buffer
	var startDateTransformed *shared.Date
	var endDateTransformed *shared.Date

	startDateToTime, _ := time.Parse("2006-01-02", "2024-04-12")
	startDateTransformed = &shared.Date{Time: startDateToTime}

	endDateToTime, _ := time.Parse("2006-01-02", "2025-03-31")
	endDateTransformed = &shared.Date{Time: endDateToTime}

	manualInvoiceInfo := &shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeS2,
		Amount:           32000,
		RaisedDate:       endDateTransformed,
		StartDate:        startDateTransformed,
		EndDate:          endDateTransformed,
		SupervisionLevel: "GENERAL",
	}
	_ = json.NewEncoder(&b).Encode(manualInvoiceInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/invoices", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{manualInvoice: manualInvoiceInfo, err: errors.New("Something is wrong")}
	server := Server{Service: mock, Validator: validator}
	server.addFeeReduction(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}
