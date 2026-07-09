package allpay

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCancelMandate_Success(t *testing.T) {
	schemeCode := "SCHEME123"

	date := time.Now().AddDate(0, 0, 1) // future date

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		expectedPath := "/AllpayApi/Customers/SCHEME123/MTIzNDU2Nzg=/Q2FuY2VsbWFu/Mandates/" + date.Format("2006-01-02")
		if r.URL.Path != expectedPath {
			t.Errorf("Unexpected URL path: got %s, want %s", r.URL.Path, expectedPath)
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
		ClosureDate: date,
		ClientDetails: ClientDetails{
			ClientReference: "12345678",
			Surname:         " Cancelman ", // whitespace should be stripped before encoding
		},
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
		ClosureDate: time.Now().AddDate(0, 0, 1),
		ClientDetails: ClientDetails{
			ClientReference: "12345678",
			Surname:         "Cancelman",
		},
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
		ClosureDate: time.Now().AddDate(0, 0, 1),
		ClientDetails: ClientDetails{
			ClientReference: "12345678",
			Surname:         "Cancelman",
		},
	})
	if err == nil {
		t.Error("Expected error due to unexpected status code")
	}
}

func TestCancelMandateWhenAlreadyCancelled_ReturnsNil(t *testing.T) {
	schemeCode := "SCHEME123"

	date := time.Now().AddDate(0, 0, 1) // future date
	alreadyCancelled := ErrorValidation{Messages: []string{"A Direct Debit Mandate was not found for this account"}}
	body, _ := json.Marshal(alreadyCancelled)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		expectedPath := "/AllpayApi/Customers/SCHEME123/MTIzNDU2Nzg=/Q2FuY2VsbWFu/Mandates/" + date.Format("2006-01-02")
		if r.URL.Path != expectedPath {
			t.Errorf("Unexpected URL path: got %s, want %s", r.URL.Path, expectedPath)
		}
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write(body)
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
		ClosureDate: date,
		ClientDetails: ClientDetails{
			ClientReference: "12345678",
			Surname:         " Cancelman ", // whitespace should be stripped before encoding
		},
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCancelMandate_ValidationErrorValidJSON(t *testing.T) {
	ve := ErrorValidation{Messages: []string{"Some generic validation message"}}
	body, _ := json.Marshal(ve)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write(body)
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
		ClosureDate: time.Now().AddDate(0, 0, 1),
		ClientDetails: ClientDetails{
			ClientReference: "12345678",
			Surname:         "Cancelman",
		},
	})

	var validationErr ErrorValidation
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ErrorValidation, got %v", err)
	}
}

func TestCancelMandate_SuccessWithoutClosureDate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/AllpayApi/Customers/SCHEME123/MTIzNDU2Nzg=/Q2FuY2VsbWFu/Mandates/" {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
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
		ClosureDate: time.Now().Truncate(24 * time.Hour),
		ClientDetails: ClientDetails{
			ClientReference: "12345678",
			Surname:         " Cancelman ",
		},
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
