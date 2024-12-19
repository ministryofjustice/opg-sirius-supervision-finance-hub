package api

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmptyAccountHolderReturnsValidationError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.SubmitDirectDebit("", "testing", "123456", "12345678")
	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"AccountHolder": map[string]string{"required": ""}}}
	assert.Equal(t, expectedError, err.(apierror.ValidationError))
}

func TestEmptyAccountNameReturnsValidationError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.SubmitDirectDebit("CLIENT", "", "123456", "12345678")
	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"AccountName": map[string]string{"required": ""}}}
	assert.Equal(t, expectedError, err.(apierror.ValidationError))
}

func TestAccountNameOver18CharactersReturnsValidationError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)
	accountNameOver18Characters := strings.Repeat("a", 19)

	err := client.SubmitDirectDebit("CLIENT", accountNameOver18Characters, "123456", "12345678")
	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"AccountName": map[string]string{"gteEighteen": ""}}}
	assert.Equal(t, expectedError, err.(apierror.ValidationError))
}

func TestEmptySortCodeReturnsValidationError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.SubmitDirectDebit("CLIENT", "account Name", "", "12345678")
	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"SortCode": map[string]string{"eqSix": ""}}}
	assert.Equal(t, expectedError, err.(apierror.ValidationError))
}

func TestSortCodeNotSixCharactersReturnsValidationError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.SubmitDirectDebit("CLIENT", "account Name", "12345", "12345678")
	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"SortCode": map[string]string{"eqSix": ""}}}
	assert.Equal(t, expectedError, err.(apierror.ValidationError))
}

func TestSortCodeSixCharactersAllZerosReturnsValidationError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.SubmitDirectDebit("CLIENT", "account Name", "000000", "12345678")
	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"SortCode": map[string]string{"eqSix": ""}}}
	assert.Equal(t, expectedError, err.(apierror.ValidationError))
}

func TestAccountNumberNotEightCharactersReturnsValidationError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.SubmitDirectDebit("CLIENT", "account Name", "123456", "123456789")
	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"AccountNumber": map[string]string{"eqEight": ""}}}
	assert.Equal(t, expectedError, err.(apierror.ValidationError))
}

func TestEmptyAccountHolderReturnsNil(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

	err := client.SubmitDirectDebit("CLIENT", "testing", "123456", "12345678")
	assert.Nil(t, err)
}
