package allpay

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCancelMandate_Success(t *testing.T) {
	schemeCode := "SCHEME123"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		today := time.Now().Format("2006-01-02")
		if r.URL.Path != "/AllpayApi/Customers/SCHEME123/MTIzNDU2Nzg=/Q2FuY2VsbWFu/Mandates/"+today {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := &Client{
		http: ts.Client(),
		Envs: Envs{
			schemeCode: schemeCode,
			apiHost:    ts.URL,
		},
	}

	err := c.CancelMandate(testContext(), &CancelMandateRequest{
		ClientReference: "12345678",
		Surname:         "Cancelman",
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCancelMandate_RequestCreationFails(t *testing.T) {
	c := &Client{
		http: http.DefaultClient,
		Envs: Envs{
			schemeCode: "SCHEME123",
			apiHost:    "http://[::1]:namedport/", // Invalid URL
		},
	}

	err := c.CancelMandate(testContext(), &CancelMandateRequest{
		ClientReference: "12345678",
		Surname:         "Cancelman",
	})
	if err == nil {
		t.Error("Expected error due to request creation failure")
	}
}

func TestCancelMandate_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	c := &Client{
		http: ts.Client(),
		Envs: Envs{
			schemeCode: "SCHEME123",
			apiHost:    ts.URL,
		},
	}

	err := c.CancelMandate(testContext(), &CancelMandateRequest{
		ClientReference: "12345678",
		Surname:         "Cancelman",
	})
	if err == nil {
		t.Error("Expected error due to unexpected status code")
	}
}
