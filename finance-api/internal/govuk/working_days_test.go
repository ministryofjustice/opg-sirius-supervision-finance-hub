package govuk

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
)

func TestIsWeekend(t *testing.T) {
	client := &Client{}

	tests := []struct {
		date     time.Time
		expected bool
	}{
		{time.Date(2025, 8, 16, 0, 0, 0, 0, time.UTC), true},  // Saturday
		{time.Date(2025, 8, 17, 0, 0, 0, 0, time.UTC), true},  // Sunday
		{time.Date(2025, 8, 18, 0, 0, 0, 0, time.UTC), false}, // Monday
	}

	for _, tt := range tests {
		result := client.isWeekend(tt.date)
		if result != tt.expected {
			t.Errorf("isWeekend(%v) = %v; want %v", tt.date, result, tt.expected)
		}
	}
}

func TestIsWorkingDay(t *testing.T) {
	client := &Client{
		caches: newCaches(),
	}
	client.caches.updateHolidays([]Holiday{{"2025-12-25"}})

	tests := []struct {
		date     time.Time
		expected bool
	}{
		{time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC), false}, // Holiday
		{time.Date(2025, 8, 16, 0, 0, 0, 0, time.UTC), false},  // Saturday
		{time.Date(2025, 8, 18, 0, 0, 0, 0, time.UTC), true},   // Monday
	}

	for _, tt := range tests {
		result := client.isWorkingDay(tt.date)
		if result != tt.expected {
			t.Errorf("isWorkingDay(%v) = %v; want %v", tt.date, result, tt.expected)
		}
	}
}

func TestAddWorkingDays(t *testing.T) {
	client := &Client{
		caches: newCaches(),
	}
	client.caches.updateHolidays([]Holiday{{"2025-12-25"}, {"2025-12-26"}})

	start := time.Date(2025, 12, 24, 0, 0, 0, 0, time.UTC) // Wednesday
	// Thursday 25th - Holiday
	// Friday 26th - Holiday
	// Saturday 27th - Weekend
	// Sunday 28th - Weekend
	// Monday 29th - Work day
	// Tuesday 30th - Work day
	// Wednesday 31st - Work day
	expected := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	result, err := client.AddWorkingDays(testContext(), start, 3)
	if err != nil {
		t.Fatalf("AddWorkingDays returned error: %v", err)
	}

	if !result.Equal(expected) {
		t.Errorf("AddWorkingDays = %v; want %v", result, expected)
	}
}

func TestAddWorkingDays_cache_refresh(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
	  	"england-and-wales": {
			"division": "england-and-wales",
			"events": [
			  {
				"title": "Christmas Day",
				"date": "2025-12-25",
				"notes": "",
				"bunting": true
			  },
			  {
				"title": "Boxing Day",
				"date": "2025-12-26",
				"notes": "",
				"bunting": true
			  }
			]
		  }
		}`))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := &Client{
		http:   ts.Client(),
		Envs:   Envs{HolidayAPIURL: ts.URL},
		caches: &Caches{holidays: cache.New(defaultExpiration, defaultExpiration)}, // no cached values or refresh trigger
	}

	start := time.Date(2025, 12, 24, 0, 0, 0, 0, time.UTC)
	expected := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	result, err := client.AddWorkingDays(testContext(), start, 3)
	if err != nil {
		t.Fatalf("AddWorkingDays returned error: %v", err)
	}

	if !result.Equal(expected) {
		t.Errorf("AddWorkingDays = %v; want %v", result, expected)
	}
}

func TestNextWorkingDayOnOrAfterX(t *testing.T) {
	client := &Client{
		caches: newCaches(),
	}
	client.caches.updateHolidays([]Holiday{{"2025-01-24"}, {"2025-01-25"}})
	X := 24

	tests := []struct {
		name     string
		date     time.Time
		expected time.Time
	}{
		{
			name:     "before X",
			date:     time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 12, X, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "on X",
			date:     time.Date(2024, 12, X, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 12, X, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "after X",
			date:     time.Date(2025, 1, 25, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2025, 2, X, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "on X with working days",
			date:     time.Date(2025, 1, X, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2025, 1, 26, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "after X on working days",
			date:     time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2025, 1, 26, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		result, err := client.NextWorkingDayOnOrAfterX(testContext(), tt.date, X)
		if err != nil {
			t.Fatalf("NextWorkingDayOnOrAfterX returned error: %v", err)
		}

		if !result.Equal(tt.expected) {
			t.Errorf("NextWorkingDayOnOrAfterX:%s = %v; want %v", tt.name, result, tt.expected)
		}
	}
}

func TestNextWorkingDayOnOrAfterX(t *testing.T) {
	client := &Client{
		caches: newCaches(),
	}
	client.caches.updateHolidays([]Holiday{{"2025-01-24"}, {"2025-01-25"}})
	X := 24

	tests := []struct {
		name     string
		date     time.Time
		expected time.Time
	}{
		{
			name:     "before X",
			date:     time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 12, X, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "on X",
			date:     time.Date(2024, 12, X, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 12, X, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "after X",
			date:     time.Date(2025, 1, 25, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2025, 2, X, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "on X with working days",
			date:     time.Date(2025, 1, X, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2025, 1, 26, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "after X on working days",
			date:     time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2025, 1, 26, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		result, err := client.NextWorkingDayOnOrAfterX(testContext(), tt.date, X)
		if err != nil {
			t.Fatalf("NextWorkingDayOnOrAfterX returned error: %v", err)
		}

		if !result.Equal(tt.expected) {
			t.Errorf("NextWorkingDayOnOrAfterX:%s = %v; want %v", tt.name, result, tt.expected)
		}
	}
}
