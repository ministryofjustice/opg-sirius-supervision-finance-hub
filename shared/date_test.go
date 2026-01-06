package shared

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

type testJsonDateStruct struct {
	TestDate Date `json:"testDate"`
}

func TestNewDate(t *testing.T) {
	want := Date{Time: time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)}
	assert.Equal(t, want, NewDate("31/12/2020"))

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewDate should panic with incorrect date format")
		}
	}()
	NewDate("12/31/2020") // wrong format should trigger a panic
}

func TestDate_Before_And_After(t *testing.T) {
	tests := []struct {
		name              string
		date1             Date
		date2             Date
		wantForBeforeTest bool
		wantForAfterTest  bool
	}{
		{
			name:              "Date1 is before Date2",
			date1:             NewDate("01/01/2020"),
			date2:             NewDate("02/01/2020"),
			wantForBeforeTest: true,
			wantForAfterTest:  false,
		},
		{
			name:              "Date1 is after Date2",
			date1:             NewDate("02/01/2020"),
			date2:             NewDate("01/01/2020"),
			wantForBeforeTest: false,
			wantForAfterTest:  true,
		},
		{
			name:              "Date1 is the same as Date2",
			date1:             NewDate("01/01/2020"),
			date2:             NewDate("01/01/2020"),
			wantForBeforeTest: false,
			wantForAfterTest:  false,
		},
		{
			name:              "Date1 is empty",
			date1:             Date{},
			date2:             NewDate("02/01/2020"),
			wantForBeforeTest: true,
			wantForAfterTest:  false,
		},
		{
			name:              "Date2 is empty",
			date1:             NewDate("01/01/2020"),
			date2:             Date{},
			wantForBeforeTest: false,
			wantForAfterTest:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.wantForBeforeTest, test.date1.Before(test.date2))
			assert.Equal(t, test.wantForAfterTest, test.date1.After(test.date2))
		})
	}
}

func TestDate_IsNull(t *testing.T) {
	tests := []struct {
		date Date
		want bool
	}{
		{
			date: NewDate("01/01/2020"),
			want: false,
		},
		{
			date: NewDate("01/01/0001"),
			want: true,
		},
		{
			date: Date{},
			want: true,
		},
	}
	for i, test := range tests {
		t.Run("Scenario "+strconv.Itoa(i+1), func(t *testing.T) {
			assert.Equal(t, test.want, test.date.IsNull())
		})
	}
}

func TestDate_isSameFinancialYear(t *testing.T) {
	type args struct {
		startDate Date
		endDate   Date
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "returns false if the start date and end date are not in the same financial year",
			args: args{
				startDate: NewDate("01/05/2024"),
				endDate:   Date{Time: time.Date(2025, 5, 02, 0, 0, 0, 0, time.UTC)},
			},
			want: false,
		},
		{
			name: "returns true if the start date and end date are in the same financial year",
			args: args{
				startDate: NewDate("01/04/2024"),
				endDate:   Date{Time: time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC)},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.args.startDate.IsSameFinancialYear(tt.args.endDate), "isSameFinancialYear(%v, %v)", tt.args.startDate, tt.args.endDate)
		})
	}
}

func TestDate_MarshalJSON(t *testing.T) {
	v := testJsonDateStruct{TestDate: NewDate("01/01/2020")}
	b, err := json.Marshal(v)
	assert.Nil(t, err)
	assert.Equal(t, `{"testDate":"01\/01\/2020"}`, string(b))
}

func TestDate_String(t *testing.T) {
	tests := []struct {
		name      string
		inputDate string
		want      string
	}{
		{
			name:      "returns correct format for slashers",
			inputDate: "01/01/2020",
			want:      "01/01/2020",
		},
		{
			name:      "returns correct format for dashers",
			inputDate: "2024-10-01",
			want:      "01/10/2024",
		},
		{
			name:      "returns correct format for date time string",
			inputDate: "2025-01-02T18:07:10+00:00",
			want:      "02/01/2025",
		},
		{
			name:      "returns nothing for default date string",
			inputDate: "01/01/0001",
			want:      "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewDate(tt.inputDate).String(), "stringToTime(%v)", tt.inputDate)
		})
	}
}

func TestDate_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		json string
		want string
	}{
		{
			json: `{"testDate":"01\/01\/2020"}`,
			want: "01/01/2020",
		},
		{
			json: `{"testDate":"01/01/2020"}`,
			want: "01/01/2020",
		},
		{
			json: `{"testDate":"2020-01-01T20:01:02+00:00"}`,
			want: "01/01/2020",
		},
	}
	for i, test := range tests {
		t.Run("Scenario "+strconv.Itoa(i+1), func(t *testing.T) {
			var v *testJsonDateStruct
			err := json.Unmarshal([]byte(test.json), &v)
			assert.Nil(t, err)
			assert.Equal(t, test.want, v.TestDate.String())
		})
	}
}

func TestDate_calculateFinanceYear(t *testing.T) {
	type args struct {
		startDate Date
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "start date is in April",
			args: args{
				startDate: NewDate("01/04/2024"),
			},
			want: "24",
		},
		{
			name: "start date and is in January",
			args: args{
				startDate: NewDate("01/01/2025"),
			},
			want: "24",
		}, {
			name: "start date and is in March",
			args: args{
				startDate: NewDate("31/03/2025"),
			},
			want: "24",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.args.startDate.CalculateFinanceYear(), "calculateFinanceYear(%v)", tt.args.startDate)
		})
	}
}

func TestDate_getCurrentDateWithoutTime(t *testing.T) {
	assert.Equal(t, time.Now().Year(), GetCurrentDateWithoutTime().Year())
	assert.Equal(t, time.Now().Month(), GetCurrentDateWithoutTime().Month())
	assert.Equal(t, time.Now().Day(), GetCurrentDateWithoutTime().Day())
	assert.Equal(t, 0, GetCurrentDateWithoutTime().Hour())
	assert.Equal(t, 0, GetCurrentDateWithoutTime().Minute())
	assert.Equal(t, 0, GetCurrentDateWithoutTime().Second())
	assert.Equal(t, 0, GetCurrentDateWithoutTime().Nanosecond())
}
