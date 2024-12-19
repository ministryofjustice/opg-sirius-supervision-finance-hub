package api

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSubmitDirectDebitValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		accountHolder string
		accountName   string
		sortCode      string
		accountNumber string
		expectedError apierror.ValidationError
	}{
		{
			name:          "Empty AccountHolder Returns ValidationError",
			accountHolder: "",
			accountName:   "testing",
			sortCode:      "123456",
			accountNumber: "12345678",
			expectedError: apierror.ValidationError{Errors: apierror.ValidationErrors{"AccountHolder": map[string]string{"required": ""}}},
		},
		{
			name:          "Empty AccountName Returns ValidationError",
			accountHolder: "CLIENT",
			accountName:   "",
			sortCode:      "123456",
			accountNumber: "12345678",
			expectedError: apierror.ValidationError{Errors: apierror.ValidationErrors{"AccountName": map[string]string{"required": ""}}},
		},
		{
			name:          "AccountName Over 18 Characters Returns ValidationError",
			accountHolder: "CLIENT",
			accountName:   strings.Repeat("a", 19),
			sortCode:      "123456",
			accountNumber: "12345678",
			expectedError: apierror.ValidationError{Errors: apierror.ValidationErrors{"AccountName": map[string]string{"gteEighteen": ""}}},
		},
		{
			name:          "Empty SortCode Returns ValidationError",
			accountHolder: "CLIENT",
			accountName:   "account Name",
			sortCode:      "",
			accountNumber: "12345678",
			expectedError: apierror.ValidationError{Errors: apierror.ValidationErrors{"SortCode": map[string]string{"eqSix": ""}}},
		},
		{
			name:          "SortCode Not Six Characters Returns ValidationError",
			accountHolder: "CLIENT",
			accountName:   "account Name",
			sortCode:      "12345",
			accountNumber: "12345678",
			expectedError: apierror.ValidationError{Errors: apierror.ValidationErrors{"SortCode": map[string]string{"eqSix": ""}}},
		},
		{
			name:          "SortCode Six Characters All Zeros Returns ValidationError",
			accountHolder: "CLIENT",
			accountName:   "account Name",
			sortCode:      "000000",
			accountNumber: "12345678",
			expectedError: apierror.ValidationError{Errors: apierror.ValidationErrors{"SortCode": map[string]string{"eqSix": ""}}},
		},
		{
			name:          "AccountNumber Not Eight Characters Returns ValidationError",
			accountHolder: "CLIENT",
			accountName:   "account Name",
			sortCode:      "123456",
			accountNumber: "123456789",
			expectedError: apierror.ValidationError{Errors: apierror.ValidationErrors{"AccountNumber": map[string]string{"eqEight": ""}}},
		},
		{
			name:          "Successful payload Returns Nil",
			accountHolder: "CLIENT",
			accountName:   "testing",
			sortCode:      "123456",
			accountNumber: "12345678",
			expectedError: apierror.ValidationError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			defer svr.Close()

			client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL)

			err := client.SubmitDirectDebit(tt.accountHolder, tt.accountName, tt.sortCode, tt.accountNumber)

			if tt.expectedError.Errors != nil {
				assert.Equal(t, tt.expectedError, err.(apierror.ValidationError))
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
