package api

import (
	"github.com/patrickmn/go-cache"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

	result, err := client.addWorkingDays(testContext(), start, 3)
	if err != nil {
		t.Fatalf("addWorkingDays returned error: %v", err)
	}

	if !result.Equal(expected) {
		t.Errorf("addWorkingDays = %v; want %v", result, expected)
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

	result, err := client.addWorkingDays(testContext(), start, 3)
	if err != nil {
		t.Fatalf("addWorkingDays returned error: %v", err)
	}

	if !result.Equal(expected) {
		t.Errorf("addWorkingDays = %v; want %v", result, expected)
	}
}
