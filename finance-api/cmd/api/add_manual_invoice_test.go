package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func TestServer_addManualInvoice(t *testing.T) {
	var b bytes.Buffer

	startDateToTime, _ := time.Parse("2006-01-02", "2024-04-12")
	endDateToTime, _ := time.Parse("2006-01-02", "2025-03-31")

	manualInvoiceInfo := &shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeS2,
		Amount:           shared.Nillable[int32]{Value: 32000, Valid: true},
		RaisedDate:       shared.Nillable[shared.Date]{Value: shared.Date{Time: endDateToTime}, Valid: true},
		StartDate:        shared.Nillable[shared.Date]{Value: shared.Date{Time: startDateToTime}, Valid: true},
		EndDate:          shared.Nillable[shared.Date]{Value: shared.Date{Time: endDateToTime}, Valid: true},
		SupervisionLevel: shared.Nillable[string]{Value: "GENERAL", Valid: true},
	}
	_ = json.NewEncoder(&b).Encode(manualInvoiceInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/invoices", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{manualInvoice: manualInvoiceInfo}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	_ = server.addManualInvoice(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)
	data, _ := io.ReadAll(res.Body)

	expected := ""

	assert.Equal(t, expected, string(data))
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestServer_addManualInvoiceNoValidationErrorsForNilFields(t *testing.T) {
	var b bytes.Buffer
	manualInvoiceInfo := &shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeAD,
		Amount:           shared.Nillable[int32]{},
		RaisedDate:       shared.Nillable[shared.Date]{},
		StartDate:        shared.Nillable[shared.Date]{},
		EndDate:          shared.Nillable[shared.Date]{},
		SupervisionLevel: shared.Nillable[string]{},
	}
	_ = json.NewEncoder(&b).Encode(manualInvoiceInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/invoices", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{manualInvoice: manualInvoiceInfo}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	_ = server.addManualInvoice(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	expected := ""

	assert.Equal(t, expected, w.Body.String())
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestServer_addManualInvoiceValidationErrors(t *testing.T) {
	var b bytes.Buffer

	manualInvoiceInfo := &shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeUnknown,
		Amount:           shared.Nillable[int32]{Valid: true},
		RaisedDate:       shared.Nillable[shared.Date]{Valid: true},
		StartDate:        shared.Nillable[shared.Date]{Valid: true},
		EndDate:          shared.Nillable[shared.Date]{Valid: true},
		SupervisionLevel: shared.Nillable[string]{Valid: true},
	}
	_ = json.NewEncoder(&b).Encode(manualInvoiceInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/invoices", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{manualInvoice: manualInvoiceInfo}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.addManualInvoice(w, req)

	expected := apierror.ValidationError{Errors: apierror.ValidationErrors{
		"Amount": {
			"nillable-int-gt": "This field Amount needs to be looked at nillable-int-gt",
		},
		"EndDate": {
			"nillable-date-required": "This field EndDate needs to be looked at nillable-date-required",
		},
		"InvoiceType": {
			"required": "This field InvoiceType needs to be looked at required",
		},
		"RaisedDate": {
			"nillable-date-required": "This field RaisedDate needs to be looked at nillable-date-required",
		},
		"StartDate": {
			"nillable-date-required": "This field StartDate needs to be looked at nillable-date-required",
		},
		"SupervisionLevel": {
			"nillable-string-oneof": "This field SupervisionLevel needs to be looked at nillable-string-oneof",
		},
	}}
	assert.Equal(t, expected, err)
}

func TestServer_addManualInvoiceValidationErrorsForAmountTooHigh(t *testing.T) {
	var b bytes.Buffer

	startDateToTime, _ := time.Parse("2006-01-02", "2024-04-12")
	endDateToTime, _ := time.Parse("2006-01-02", "2025-03-31")

	manualInvoiceInfo := &shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeS2,
		Amount:           shared.Nillable[int32]{Value: 320000, Valid: true},
		RaisedDate:       shared.Nillable[shared.Date]{Value: shared.Date{Time: endDateToTime}, Valid: true},
		StartDate:        shared.Nillable[shared.Date]{Value: shared.Date{Time: startDateToTime}, Valid: true},
		EndDate:          shared.Nillable[shared.Date]{Value: shared.Date{Time: endDateToTime}, Valid: true},
		SupervisionLevel: shared.Nillable[string]{Value: "GENERAL", Valid: true},
	}
	_ = json.NewEncoder(&b).Encode(manualInvoiceInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/invoices", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{manualInvoice: manualInvoiceInfo}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.addManualInvoice(w, req)

	expected := apierror.ValidationError{Errors: apierror.ValidationErrors{
		"Amount": {
			"nillable-int-lte": "This field Amount needs to be looked at nillable-int-lte",
		},
	}}
	assert.Equal(t, expected, err)
}

func TestServer_addManualInvoiceDateErrors(t *testing.T) {
	var b bytes.Buffer

	startDateToTime, _ := time.Parse("2006-01-02", "2024-04-12")
	endDateToTime, _ := time.Parse("2006-01-02", "2025-03-31")

	manualInvoiceInfo := &shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeS2,
		Amount:           shared.Nillable[int32]{Value: 320000, Valid: true},
		RaisedDate:       shared.Nillable[shared.Date]{Value: shared.Date{Time: endDateToTime}, Valid: true},
		StartDate:        shared.Nillable[shared.Date]{Value: shared.Date{Time: startDateToTime}, Valid: true},
		EndDate:          shared.Nillable[shared.Date]{Value: shared.Date{Time: endDateToTime}, Valid: true},
		SupervisionLevel: shared.Nillable[string]{Value: "GENERAL", Valid: true},
	}
	_ = json.NewEncoder(&b).Encode(manualInvoiceInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/invoices", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{manualInvoice: manualInvoiceInfo, errs: map[string]error{"AddFeeReduction": apierror.BadRequestsError([]string{"RaisedDateForAnInvoice", "StartDate", "EndDate"})}}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.addFeeReduction(w, req)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	var expErr apierror.ValidationError
	assert.ErrorAs(t, err, &expErr)
}

func TestServer_addManualInvoice422Error(t *testing.T) {
	var b bytes.Buffer

	startDateToTime, _ := time.Parse("2006-01-02", "2024-04-12")
	endDateToTime, _ := time.Parse("2006-01-02", "2025-03-31")

	manualInvoiceInfo := &shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeS2,
		Amount:           shared.Nillable[int32]{Value: 32000, Valid: true},
		RaisedDate:       shared.Nillable[shared.Date]{Value: shared.Date{Time: endDateToTime}, Valid: true},
		StartDate:        shared.Nillable[shared.Date]{Value: shared.Date{Time: startDateToTime}, Valid: true},
		EndDate:          shared.Nillable[shared.Date]{Value: shared.Date{Time: endDateToTime}, Valid: true},
		SupervisionLevel: shared.Nillable[string]{Value: "GENERAL", Valid: true},
	}
	_ = json.NewEncoder(&b).Encode(manualInvoiceInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/invoices", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{manualInvoice: manualInvoiceInfo, errs: map[string]error{"AddFeeReduction": errors.New("something is wrong")}}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.addFeeReduction(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	var expErr apierror.ValidationError
	assert.ErrorAs(t, err, &expErr)
}
