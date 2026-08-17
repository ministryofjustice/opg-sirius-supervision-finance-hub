package allpay

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateMandate_Success(t *testing.T) {
	schemeCode := "SCHEME123"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/AllpayApi/Customers/SCHEME123/Mandates/Create" {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var req CreateMandateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("Invalid JSON body: %v", err)
		}
		if req.Customer.SchemeCode != schemeCode {
			t.Errorf("Expected schemeCode to be added, got %s", req.Customer.SchemeCode)
		}
		if req.Customer.Surname != "Doe" {
			t.Errorf("Expected surname to be sent as provided, got %s", req.Customer.Surname)
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

	err := c.CreateMandate(testContext(), &CreateMandateInput{
		Customer: Customer{
			ClientDetails: ClientDetails{
				ClientReference: "REF123",
				Surname:         "Doe",
			},
			Address: Address{
				Line1:    "1 Street",
				Town:     "Townville",
				PostCode: "TO1 2ST",
			},
		},
		BankDetails: BankDetails{
			AccountName:   "Donny Doe",
			SortCode:      "11-22-33",
			AccountNumber: "12345678",
		},
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCreateMandate_DataFormatting(t *testing.T) {
	tests := []struct {
		name             string
		input            CreateMandateInput
		expectedLine1    string
		expectedTown     string
		expectedPostCode string
	}{
		{
			name: "trims whitespace from address fields",
			input: CreateMandateInput{Customer: Customer{
				Address: Address{Line1: "  1 High Street  ", Town: "  London  ", PostCode: "  SW1A 1AA  "},
			}},
			expectedLine1:    "1 High Street",
			expectedTown:     "London",
			expectedPostCode: "SW1A 1AA",
		},
		{
			name: "truncates address line1 to 40 characters",
			input: CreateMandateInput{Customer: Customer{
				Address: Address{Line1: "123 This Is A Very Long Street Name That Exceeds Forty Characters"},
			}},
			expectedLine1: "123 This Is A Very Long Street Name That",
		},
		{
			name: "truncates town to 40 characters",
			input: CreateMandateInput{Customer: Customer{
				Address: Address{Town: "A Very Long Town Name That Clearly Exceeds The Forty Character Limit"},
			}},
			expectedTown: "A Very Long Town Name That Clearly Excee",
		},
		{
			name: "truncates postcode to 8 characters",
			input: CreateMandateInput{Customer: Customer{
				Address: Address{PostCode: "DA12 25TX"},
			}},
			expectedPostCode: "DA12 25T",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured CreateMandateRequest

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &captured)
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			c := &Client{
				http: ts.Client(),
				Envs: Envs{apiHost: ts.URL, schemeCode: "SCHEME"},
			}

			_ = c.CreateMandate(testContext(), &tt.input)

			if tt.expectedLine1 != "" && captured.Customer.Address.Line1 != tt.expectedLine1 {
				t.Errorf("Line1: expected %q, got %q", tt.expectedLine1, captured.Customer.Address.Line1)
			}
			if tt.expectedTown != "" && captured.Customer.Address.Town != tt.expectedTown {
				t.Errorf("Town: expected %q, got %q", tt.expectedTown, captured.Customer.Address.Town)
			}
			if tt.expectedPostCode != "" && captured.Customer.Address.PostCode != tt.expectedPostCode {
				t.Errorf("PostCode: expected %q, got %q", tt.expectedPostCode, captured.Customer.Address.PostCode)
			}
		})
	}
}

func TestCreateMandate_RequestCreationFails(t *testing.T) {
	c := &Client{
		http: http.DefaultClient,
		Envs: Envs{
			schemeCode: "SCHEME123",
			apiHost:    "http://[::1]:namedport/", // Invalid URL
		},
	}

	err := c.CreateMandate(testContext(), &CreateMandateInput{})
	if err == nil {
		t.Error("Expected error due to request creation failure")
	}
}

func TestCreateMandate_UnexpectedStatus(t *testing.T) {
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

	err := c.CreateMandate(testContext(), &CreateMandateInput{})
	if err == nil {
		t.Error("Expected error due to unexpected status code")
	}
}

func TestCreateMandate_ValidationErrorValidJSON(t *testing.T) {
	ve := ErrorValidation{Messages: []string{"Invalid data"}}
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

	err := c.CreateMandate(testContext(), &CreateMandateInput{})
	var validationErr ErrorValidation
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ErrorValidation, got %v", err)
	}
}

func TestCreateMandate_AlreadyExistsWithoutScheduleTreatedAsSuccess(t *testing.T) {
	ve := ErrorValidation{Messages: []string{"This direct debit reference is already being used"}}
	body, _ := json.Marshal(ve)
	mandateCalls := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/AllpayApi/Customers/SCHEME123/Mandates/Create" {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
		mandateCalls++
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

	err := c.CreateMandate(testContext(), &CreateMandateInput{
		Customer: Customer{
			ClientDetails: ClientDetails{ClientReference: "REF123", Surname: "Doe"},
		},
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if mandateCalls != 1 {
		t.Errorf("Expected 1 mandate call, got %d", mandateCalls)
	}
}

func TestCreateMandate_AlreadyExistsWithScheduleCreatesSchedule(t *testing.T) {
	alreadyExists := ErrorValidation{Messages: []string{"This direct debit reference is already being used"}}
	alreadyExistsBody, _ := json.Marshal(alreadyExists)
	mandateCalls := 0
	scheduleCalls := 0

	collectionDate := time.Date(2026, time.July, 15, 0, 0, 0, 0, time.UTC)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/AllpayApi/Customers/SCHEME123/Mandates/Create":
			mandateCalls++
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = w.Write(alreadyExistsBody)
		case "/AllpayApi/Customers/SCHEME123/UkVGMTIz/RG9l/Mandates":
			scheduleCalls++
			body, _ := io.ReadAll(r.Body)
			var req createScheduleRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Errorf("Invalid schedule JSON body: %v", err)
			}
			expected := createScheduleRequest{
				Schedules: []Schedule{
					{
						Date:          collectionDate.Format("2006-01-02"),
						Amount:        1200,
						Frequency:     "1",
						TotalPayments: 1,
					},
				},
			}
			if len(req.Schedules) != 1 || req.Schedules[0] != expected.Schedules[0] {
				t.Errorf("Unexpected schedule payload: %+v", req)
			}
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	c := &Client{
		http: ts.Client(),
		Envs: Envs{
			schemeCode: "SCHEME123",
			apiHost:    ts.URL,
		},
	}

	err := c.CreateMandate(testContext(), &CreateMandateInput{
		Customer: Customer{
			ClientDetails: ClientDetails{
				ClientReference: "REF123",
				Surname:         "Doe",
			},
		},
		Schedule: &ScheduleInput{
			Date:   collectionDate,
			Amount: 1200,
		},
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if mandateCalls != 1 {
		t.Errorf("Expected 1 mandate call, got %d", mandateCalls)
	}
	if scheduleCalls != 1 {
		t.Errorf("Expected 1 schedule call, got %d", scheduleCalls)
	}
}
