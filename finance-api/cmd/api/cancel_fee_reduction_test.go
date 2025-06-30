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
	"testing"
)

func TestServer_cancelFeeReduction(t *testing.T) {
	var b bytes.Buffer

	cancelFeeReductionInfo := &shared.CancelFeeReduction{
		CancellationReason: "Cancelling a remission reduction",
	}
	_ = json.NewEncoder(&b).Encode(cancelFeeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions/1/cancel", &b)
	req.SetPathValue("clientId", "1")
	req.SetPathValue("feeReductionId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{cancelFeeReduction: cancelFeeReductionInfo}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	_ = server.cancelFeeReduction(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := ""

	assert.Equal(t, expected, w.Body.String())
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestServer_cancelFeeReductionsValidationErrors(t *testing.T) {
	var b bytes.Buffer
	cancelFeeReductionInfo := &shared.CancelFeeReduction{
		CancellationReason: "",
	}
	_ = json.NewEncoder(&b).Encode(cancelFeeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions/1/cancel", &b)
	req.SetPathValue("clientId", "1")
	req.SetPathValue("feeReductionId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{cancelFeeReduction: cancelFeeReductionInfo}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.cancelFeeReduction(w, req)

	expected := apierror.ValidationError{Errors: apierror.ValidationErrors{
		"CancellationReason": {
			"required": "This field CancellationReason needs to be looked at required",
		},
	},
	}
	assert.Equal(t, expected, err)
}

func TestServer_cancelFeeReductions500Error(t *testing.T) {
	var b bytes.Buffer
	cancelFeeReductionInfo := &shared.CancelFeeReduction{
		CancellationReason: "Adding a remission reduction",
	}
	_ = json.NewEncoder(&b).Encode(cancelFeeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions/1/cancel", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{cancelFeeReduction: cancelFeeReductionInfo, err: errors.New("Something is wrong")}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.cancelFeeReduction(w, req)

	assert.Error(t, err)
}
