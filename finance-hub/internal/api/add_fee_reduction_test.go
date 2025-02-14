package api

import (
	"bytes"
	"encoding/json"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddFeeReduction(t *testing.T) {
	mockClient := SetUpTest()
	mockJWT := mockJWTClient{}
	client := NewClient(mockClient, &mockJWT, Envs{"http://localhost:3000", ""})

	json := `{
			"id":                "1",
			"feeType":           "remission",
			"startYear":         "2025",
			"lengthOfAward":     3,
			"dateReceived":      "15/02/2024",
			"notes": "Fee remission note for one award",
        }`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 201,
			Body:       r,
		}, nil
	}

	err := client.AddFeeReduction(testContext(), 1, "remission", "2025", "3", "15/02/2024", "Fee remission note for one award")
	assert.Equal(t, nil, err)
}

func TestAddFeeReductionUnauthorised(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})

	err := client.AddFeeReduction(testContext(), 1, "remission", "2025", "3", "15/02/2024", "Fee remission note for one award")

	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
}

func TestFeeReductionReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})

	err := client.AddFeeReduction(testContext(), 1, "remission", "2025", "3", "15/02/2024", "Fee remission note for one award")
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/fee-reductions",
		Method: http.MethodPost,
	}, err)
}

func TestFeeReductionReturnsBadRequestError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})

	err := client.AddFeeReduction(testContext(), 1, "remission", "2025", "3", "15/02/2024", "Fee remission note for one award")
	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"Overlap": map[string]string{"start-or-end-date": ""}}}
	assert.Equal(t, expectedError, err)
}

func TestFeeReductionReturnsValidationError(t *testing.T) {
	validationErrors := apierror.ValidationError{
		Errors: map[string]map[string]string{
			"DateReceived": {
				"date-in-the-past": "This field DateReceived needs to be looked at date-in-the-past",
			},
		},
	}
	responseBody, _ := json.Marshal(validationErrors)
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write(responseBody)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})

	err := client.AddFeeReduction(testContext(), 0, "", "", "", "", "")
	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"DateReceived": map[string]string{"date-in-the-past": "This field DateReceived needs to be looked at date-in-the-past"}}}
	assert.Equal(t, expectedError, err.(apierror.ValidationError))
}
