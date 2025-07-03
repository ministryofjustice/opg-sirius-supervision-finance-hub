package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestServer_addFeeReductions(t *testing.T) {
	var b bytes.Buffer
	var dateReceivedTransformed *shared.Date

	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)
	dateReceivedTransformed = &shared.Date{Time: date}

	feeReductionInfo := &shared.AddFeeReduction{
		FeeType:       shared.FeeReductionTypeRemission,
		StartYear:     "2022",
		LengthOfAward: 1,
		DateReceived:  dateReceivedTransformed,
		Notes:         "Adding a remission reduction",
	}
	_ = json.NewEncoder(&b).Encode(feeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{feeReduction: feeReductionInfo}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	_ = server.addFeeReduction(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := ""

	assert.Equal(t, expected, w.Body.String())
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestServer_addFeeReductionsValidationErrors(t *testing.T) {
	var b bytes.Buffer
	feeReductionInfo := &shared.AddFeeReduction{
		FeeType:       shared.FeeReductionTypeUnknown,
		StartYear:     "",
		LengthOfAward: 0,
		DateReceived:  nil,
		Notes:         "",
	}
	_ = json.NewEncoder(&b).Encode(feeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{feeReduction: feeReductionInfo}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.addFeeReduction(w, req)

	expected := apierror.ValidationError{Errors: apierror.ValidationErrors{
		"DateReceived": {
			"required": "This field DateReceived needs to be looked at required",
		},
		"FeeType": {
			"required": "This field FeeType needs to be looked at required",
		},
		"LengthOfAward": {
			"required": "This field LengthOfAward needs to be looked at required",
		},
		"Notes": {
			"required": "This field Notes needs to be looked at required",
		},
		"StartYear": {
			"required": "This field StartYear needs to be looked at required",
		},
	}}
	assert.Equal(t, expected, err)

}

func TestServer_addFeeReductionsValidationErrorsForThousandCharacters(t *testing.T) {
	var b bytes.Buffer
	var dateReceivedTransformed *shared.Date

	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)
	dateReceivedTransformed = &shared.Date{Time: date}
	feeReductionInfo := &shared.AddFeeReduction{
		FeeType:       shared.FeeReductionTypeRemission,
		StartYear:     "2024",
		LengthOfAward: 1,
		DateReceived:  dateReceivedTransformed,
		Notes:         strings.Repeat("A", 1001),
	}
	_ = json.NewEncoder(&b).Encode(feeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{feeReduction: feeReductionInfo}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.addFeeReduction(w, req)

	expected := apierror.ValidationError{Errors: apierror.ValidationErrors{
		"Notes": {
			"thousand-character-limit": "This field Notes needs to be looked at thousand-character-limit",
		},
	}}
	assert.Equal(t, expected, err)
}

func TestServer_addFeeReductionsOverlapError(t *testing.T) {
	var b bytes.Buffer
	var dateReceivedTransformed *shared.Date

	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)
	dateReceivedTransformed = &shared.Date{Time: date}
	feeReductionInfo := &shared.AddFeeReduction{
		FeeType:       shared.FeeReductionTypeRemission,
		StartYear:     "2022",
		LengthOfAward: 1,
		DateReceived:  dateReceivedTransformed,
		Notes:         "Adding a remission reduction",
	}
	_ = json.NewEncoder(&b).Encode(feeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{feeReduction: feeReductionInfo, err: apierror.BadRequest{Reason: "overlap"}}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.addFeeReduction(w, req)

	var e apierror.BadRequest
	assert.ErrorAs(t, err, &e)
}

func TestServer_addFeeReductions500Error(t *testing.T) {
	var b bytes.Buffer
	var dateReceivedTransformed *shared.Date

	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)
	dateReceivedTransformed = &shared.Date{Time: date}
	feeReductionInfo := &shared.AddFeeReduction{
		FeeType:       shared.FeeReductionTypeRemission,
		StartYear:     "2022",
		LengthOfAward: 1,
		DateReceived:  dateReceivedTransformed,
		Notes:         "Adding a remission reduction",
	}
	_ = json.NewEncoder(&b).Encode(feeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{feeReduction: feeReductionInfo, err: errors.New("something is wrong")}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.addFeeReduction(w, req)
	assert.Error(t, err)
}
