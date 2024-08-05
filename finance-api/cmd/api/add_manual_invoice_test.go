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

	startDateToTime, _ := time.Parse("2006-01-02", "2024-04-12")
	endDateToTime, _ := time.Parse("2006-01-02", "2025-03-31")

	manualInvoiceInfo := &shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeS2,
		Amount:           shared.NillableInt{Value: 32000, Valid: true},
		RaisedDate:       shared.NillableDate{Value: shared.Date{Time: endDateToTime}, Valid: true},
		StartDate:        shared.NillableDate{Value: shared.Date{Time: startDateToTime}, Valid: true},
		EndDate:          shared.NillableDate{Value: shared.Date{Time: endDateToTime}, Valid: true},
		SupervisionLevel: shared.NillableString{Value: "GENERAL", Valid: true},
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

func TestServer_addManualInvoiceNoValidationErrorsForNilFields(t *testing.T) {
	var b bytes.Buffer
	manualInvoiceInfo := &shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeAD,
		Amount:           shared.NillableInt{},
		RaisedDate:       shared.NillableDate{},
		StartDate:        shared.NillableDate{},
		EndDate:          shared.NillableDate{},
		SupervisionLevel: shared.NillableString{},
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
		Amount:           shared.NillableInt{Valid: true},
		RaisedDate:       shared.NillableDate{Valid: true},
		RaisedYear:       shared.NillableInt{Valid: true},
		StartDate:        shared.NillableDate{Valid: true},
		EndDate:          shared.NillableDate{Valid: true},
		SupervisionLevel: shared.NillableString{Valid: true},
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
{"Message":"","validation_errors":{"Amount":{"nillable-int-required":"This field Amount needs to be looked at nillable-int-required"},"EndDate":{"nillable-date-required":"This field EndDate needs to be looked at nillable-date-required"},"InvoiceType":{"required":"This field InvoiceType needs to be looked at required"},"RaisedDate":{"nillable-date-required":"This field RaisedDate needs to be looked at nillable-date-required"},"RaisedYear":{"nillable-int-required":"This field RaisedYear needs to be looked at nillable-int-required"},"StartDate":{"nillable-date-required":"This field StartDate needs to be looked at nillable-date-required"},"SupervisionLevel":{"nillable-string-oneof":"This field SupervisionLevel needs to be looked at nillable-string-oneof"}}}`

	assert.Equal(t, expected, string(data))
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestServer_addManualInvoiceValidationErrorsForAmountTooHigh(t *testing.T) {
	var b bytes.Buffer

	startDateToTime, _ := time.Parse("2006-01-02", "2024-04-12")
	endDateToTime, _ := time.Parse("2006-01-02", "2025-03-31")

	manualInvoiceInfo := &shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeS2,
		Amount:           shared.NillableInt{Value: 320000, Valid: true},
		RaisedDate:       shared.NillableDate{Value: shared.Date{Time: endDateToTime}, Valid: true},
		StartDate:        shared.NillableDate{Value: shared.Date{Time: startDateToTime}, Valid: true},
		EndDate:          shared.NillableDate{Value: shared.Date{Time: endDateToTime}, Valid: true},
		SupervisionLevel: shared.NillableString{Value: "GENERAL", Valid: true},
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
{"Message":"","validation_errors":{"Amount":{"nillable-int-lte":"This field Amount needs to be looked at nillable-int-lte"}}}`

	assert.Equal(t, expected, string(data))
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestServer_addManualInvoiceDateErrors(t *testing.T) {
	var b bytes.Buffer

	startDateToTime, _ := time.Parse("2006-01-02", "2024-04-12")
	endDateToTime, _ := time.Parse("2006-01-02", "2025-03-31")

	manualInvoiceInfo := &shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeS2,
		Amount:           shared.NillableInt{Value: 320000, Valid: true},
		RaisedDate:       shared.NillableDate{Value: shared.Date{Time: endDateToTime}, Valid: true},
		StartDate:        shared.NillableDate{Value: shared.Date{Time: startDateToTime}, Valid: true},
		EndDate:          shared.NillableDate{Value: shared.Date{Time: endDateToTime}, Valid: true},
		SupervisionLevel: shared.NillableString{Value: "GENERAL", Valid: true},
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

	startDateToTime, _ := time.Parse("2006-01-02", "2024-04-12")
	endDateToTime, _ := time.Parse("2006-01-02", "2025-03-31")

	manualInvoiceInfo := &shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeS2,
		Amount:           shared.NillableInt{Value: 32000, Valid: true},
		RaisedDate:       shared.NillableDate{Value: shared.Date{Time: endDateToTime}, Valid: true},
		StartDate:        shared.NillableDate{Value: shared.Date{Time: startDateToTime}, Valid: true},
		EndDate:          shared.NillableDate{Value: shared.Date{Time: endDateToTime}, Valid: true},
		SupervisionLevel: shared.NillableString{Value: "GENERAL", Valid: true},
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
