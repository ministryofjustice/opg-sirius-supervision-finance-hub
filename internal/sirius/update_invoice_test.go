package sirius

import (
	"bytes"
	"github.com/opg-sirius-finance-hub/internal/mocks"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateInvoice(t *testing.T) {
	logger, mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", logger)

	json := `{
            "invoiceType": "writeOff",
            "notes": "notes here"
        }`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 201,
			Body:       r,
		}, nil
	}

	err := client.UpdateInvoice(getContext(nil), 2, 4, "writeOff", "notes here")
	assert.Equal(t, nil, err)
}

//func TestGetPersonDetailsReturnsUnauthorisedClientError(t *testing.T) {
//	logger, _ := SetUpTest()
//	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusUnauthorized)
//	}))
//	defer svr.Close()
//
//	client, _ := NewApiClient(http.DefaultClient, svr.URL, logger)
//	_, err := client.GetPersonDetails(getContext(nil), 2)
//	assert.Equal(t, ErrUnauthorized, err)
//}

//func TestPersonDetailsReturns500Error(t *testing.T) {
//	logger, _ := SetUpTest()
//	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusInternalServerError)
//	}))
//	defer svr.Close()
//
//	client, _ := NewApiClient(http.DefaultClient, svr.URL, logger)
//
//	_, err := client.GetPersonDetails(getContext(nil), 1)
//	assert.Equal(t, StatusError{
//		Code:   http.StatusInternalServerError,
//		URL:    svr.URL + "/api/v1/clients/1",
//		Method: http.MethodGet,
//	}, err)
//}
