package allpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetchFailedPayments_Success(t *testing.T) {
	schemeCode := "SCHEME123"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/AllpayApi/Customers/SCHEME123/Mandates/FailedPayments/2025-05-10/2025-05-20") {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		page := r.URL.Path[len(r.URL.Path)-1:]
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{
		  "FailedPayments": [
			{
			  "Amount": 12345,
			  "ClientReference": "ABC123XYZ",
			  "CollectionDate": "20/05/2016 00:00:00",
			  "IsRepresented": true,
			  "LastName": "Jones",
			  "Line1": "Flat 2",
			  "ProcessedDate": "20/05/2016 00:00:00",
			  "ReasonCode": "BACS 0 : Refer to payer",
			  "SchemeCode": "OPGB"
			}
		  ],
		  "TotalRecords": 2
		}`))
		case "2":
			_, _ = w.Write([]byte(`{
		  "FailedPayments": [
			{
			  "Amount": 54321,
			  "ClientReference": "ZYX321CBA",
			  "CollectionDate": "20/05/2016 00:00:00",
			  "IsRepresented": true,
			  "LastName": "Senoj",
			  "Line1": "Flat 2",
			  "ProcessedDate": "20/05/2016 00:00:00",
			  "ReasonCode": "BACS 0 : Refer to payer",
			  "SchemeCode": "OPGB"
			}
		  ],
		  "TotalRecords": 2
		}`))
		case "3":
			_, _ = w.Write([]byte(`{
		  "FailedPayments": [],
		  "TotalRecords": 2
		}`))
		default:
			assert.Fail(t, fmt.Sprintf("page %s should not have been called", page))
		}

	}))
	defer ts.Close()

	c := &Client{
		http: ts.Client(),
		Envs: Envs{
			schemeCode: schemeCode,
			apiHost:    ts.URL,
		},
	}

	from, _ := time.Parse("2006-01-02", "2025-05-10")
	to, _ := time.Parse("2006-01-02", "2025-05-20")

	fp, err := c.FetchFailedPayments(testContext(), FetchFailedPaymentsInput{
		To:   to,
		From: from,
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := FailedPayments{
		{
			Amount:          12345,
			ClientReference: "ABC123XYZ",
			CollectionDate:  "20/05/2016 00:00:00",
			ProcessedDate:   "20/05/2016 00:00:00",
			ReasonCode:      "BACS 0 : Refer to payer",
		},
		{
			Amount:          54321,
			ClientReference: "ZYX321CBA",
			CollectionDate:  "20/05/2016 00:00:00",
			ProcessedDate:   "20/05/2016 00:00:00",
			ReasonCode:      "BACS 0 : Refer to payer",
		},
	}
	assert.Equal(t, expected, fp)
}

func TestFetchFailedPayments_SuccessNoPayments(t *testing.T) {
	schemeCode := "SCHEME123"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/AllpayApi/Customers/SCHEME123/Mandates/FailedPayments") {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		page := r.URL.Path[len(r.URL.Path)-1:]
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{
		  "FailedPayments": [],
		  "TotalRecords": 0
		}`))
		default:
			assert.Fail(t, fmt.Sprintf("page %s should not have been called", page))
		}

	}))
	defer ts.Close()

	c := &Client{
		http: ts.Client(),
		Envs: Envs{
			schemeCode: schemeCode,
			apiHost:    ts.URL,
		},
	}

	fp, err := c.FetchFailedPayments(testContext(), FetchFailedPaymentsInput{
		To:   time.Time{},
		From: time.Time{},
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := FailedPayments{}
	assert.Equal(t, expected, fp)
}

func TestFetchFailedPayments_RequestCreationFails(t *testing.T) {
	c := &Client{
		http: http.DefaultClient,
		Envs: Envs{
			schemeCode: "SCHEME123",
			apiHost:    "http://[::1]:namedport/", // Invalid URL
		},
	}

	_, err := c.FetchFailedPayments(testContext(), FetchFailedPaymentsInput{
		To:   time.Time{},
		From: time.Time{},
	})
	if err == nil {
		t.Error("Expected error due to request creation failure")
	}
}

func TestFetchFailedPayments_UnexpectedStatus(t *testing.T) {
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

	_, err := c.FetchFailedPayments(testContext(), FetchFailedPaymentsInput{
		To:   time.Time{},
		From: time.Time{},
	})
	if err == nil {
		t.Error("Expected error due to unexpected status code")
	}
}
