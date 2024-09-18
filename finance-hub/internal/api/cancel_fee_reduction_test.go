package api

import (
	"bytes"
	"encoding/json"
	"github.com/opg-sirius-finance-hub/apierror"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCancelFeeReduction(t *testing.T) {
	mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "")

	json := `{
			"notes": "Fee remission note for cancelling",
        }`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 201,
			Body:       r,
		}, nil
	}

	err := client.CancelFeeReduction(getContext(nil), 1, 1, "Fee remission note for one award")
	assert.Equal(t, nil, err)
}

func TestCancelFeeReductionUnauthorised(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.CancelFeeReduction(getContext(nil), 1, 1, "Fee remission note for one award")

	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
}

func TestCancelFeeReductionReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.CancelFeeReduction(getContext(nil), 1, 1, "Fee remission note for one award")
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/fee-reductions/1/cancel",
		Method: http.MethodPut,
	}, err)
}

func TestCancelFeeReductionReturnsValidationError(t *testing.T) {
	validationErrors := apierror.ValidationError{
		Errors: map[string]map[string]string{
			"CancelFeeReductionNotes": {
				"required": "This field CancelFeeReductionNotes needs to be looked at required",
			},
		},
	}
	responseBody, _ := json.Marshal(validationErrors)
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write(responseBody)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.CancelFeeReduction(getContext(nil), 0, 0, "")
	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"CancelFeeReductionNotes": map[string]string{"required": "This field CancelFeeReductionNotes needs to be looked at required"}}}
	assert.Equal(t, expectedError, err.(apierror.ValidationError))
}
