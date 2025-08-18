package allpay

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestModulusCheck(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.RequestURI() != "/AllpayApi/BankAccounts?sortcode=11-22-33&accountnumber=12345678" {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
            "Valid": true,
            "DirectDebitCapable": true
        }`))
	}))
	defer ts.Close()

	client := NewClient(ts.Client(), ts.URL, "test123", "TEST")

	err := client.ModulusCheck(testContext(), "11-22-33", "12345678")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestModulusCheck_invalid(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
            "Valid": false,
            "DirectDebitCapable": false
        }`))
	}))
	defer ts.Close()

	client := NewClient(ts.Client(), ts.URL, "test123", "TEST")

	err := client.ModulusCheck(testContext(), "11-22-33", "12345678")
	assert.Equal(t, ErrorModulusCheckFailed{}, err)
}

// if the sort code doesn't exist, AllPay still returns valid, but "DirectDebitCapable" will be false
func TestModulusCheck_sortCodeDoesNotExist(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
            "Valid": true,
            "DirectDebitCapable": false
        }`))
	}))
	defer ts.Close()

	client := NewClient(ts.Client(), ts.URL, "test123", "TEST")

	err := client.ModulusCheck(testContext(), "11-22-33", "12345678")
	assert.Equal(t, ErrorModulusCheckFailed{}, err)
}

func TestModulusCheck_apiError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := NewClient(ts.Client(), ts.URL, "test123", "TEST")

	err := client.ModulusCheck(testContext(), "11-22-33", "12345678")
	assert.Equal(t, ErrorAPI{}, err)
}
