package allpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchMandate_Success(t *testing.T) {
	schemeCode := "SCHEME123"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/AllpayApi/Customers/SCHEME123/UkVGMTIz/RG9l") {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)

		switch r.URL.Path {
		case "/AllpayApi/Customers/SCHEME123/UkVGMTIz/RG9l":
			_, _ = w.Write([]byte(`{
		  "FetchMandateScheduleData": [
			{
			  "ClientReference": "REF123",
			  "LastName": "Doe",
			  "Status": "Active"
			}
		  ],
		  "TotalRecords": 1
		}`))
		default:
			assert.Fail(t, fmt.Sprintf("unexpected URL path: %s", r.URL.Path))
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

	out, err := c.FetchMandate(testContext(), FetchMandateScheduleInput{
		ClientDetails: ClientDetails{
			ClientReference: "REF123",
			Surname:         " Doe ",
		},
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := &FetchMandateScheduleOutput{
		FetchMandateScheduleDataType: FetchMandateScheduleDataType{
			{
				ClientReference: "REF123",
				LastName:        "Doe",
				Status:          "Active",
			},
		},
		TotalRecords: 1,
	}

	assert.Equal(t, expected, out)
}

func TestFetchSchedule_Success(t *testing.T) {
	schemeCode := "SCHEME123"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/AllpayApi/Customers/SCHEME123/UkVGMTIz/RG9l/Mandates/Schedule") {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)

		switch r.URL.Path {
		case "/AllpayApi/Customers/SCHEME123/UkVGMTIz/RG9l/Mandates/Schedule":
			_, _ = w.Write([]byte(`{
		  "FetchMandateScheduleData": [
			{
			  "Amount": 12345,
			  "ClientReference": "REF123",
			  "LastName": "Doe",
			  "ScheduleDate": "2026-07-24",
			  "Status": "Active"
			}
		  ],
		  "TotalRecords": 1
		}`))
		default:
			assert.Fail(t, fmt.Sprintf("unexpected URL path: %s", r.URL.Path))
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

	out, err := c.FetchSchedule(testContext(), FetchMandateScheduleInput{
		ClientDetails: ClientDetails{
			ClientReference: "REF123",
			Surname:         "Doe",
		},
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := &FetchMandateScheduleOutput{
		FetchMandateScheduleDataType: FetchMandateScheduleDataType{
			{
				Amount:          12345,
				ClientReference: "REF123",
				LastName:        "Doe",
				ScheduleDate:    "2026-07-24",
				Status:          "Active",
			},
		},
		TotalRecords: 1,
	}
	assert.Equal(t, expected, out)
}

func TestFetchSchedule_MissingScheduleReturnsEmptyOutput(t *testing.T) {
	schemeCode := "SCHEME123"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/AllpayApi/Customers/SCHEME123/UkVGMTIz/RG9l/Mandates/Schedule") {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)

		switch r.URL.Path {
		case "/AllpayApi/Customers/SCHEME123/UkVGMTIz/RG9l/Mandates/Schedule":
			_, _ = w.Write([]byte(`{
		  "FetchMandateScheduleData": [],
		  "TotalRecords": 0
		}`))
		default:
			assert.Fail(t, fmt.Sprintf("unexpected URL path: %s", r.URL.Path))
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

	out, err := c.FetchSchedule(testContext(), FetchMandateScheduleInput{
		ClientDetails: ClientDetails{
			ClientReference: "REF123",
			Surname:         "Doe",
		},
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := &FetchMandateScheduleOutput{
		FetchMandateScheduleDataType: FetchMandateScheduleDataType{},
		TotalRecords:                 0,
	}
	assert.Equal(t, expected, out)
}
