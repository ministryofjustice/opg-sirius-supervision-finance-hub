package shared

import (
	"strings"
	"time"
)

type Date struct {
	Time time.Time
}

func NewDate(d string) Date {
	t, err := stringToTime(d)
	if err != nil {
		panic(err)
	}

	return Date{Time: t}
}

func (d Date) Before(d2 Date) bool {
	return d.Time.Before(d2.Time)
}

func (d Date) After(d2 Date) bool {
	return d.Time.After(d2.Time)
}

func (d Date) String() string {
	if d.IsNull() {
		return ""
	}
	return d.Time.Format("02/01/2006")
}

func (d Date) IsSameFinancialYear(d1 *Date) bool {
	var financialYearOneStartYear int
	if d.Time.Month() < time.April {
		financialYearOneStartYear = d.Time.Year() - 1
	} else {
		financialYearOneStartYear = d.Time.Year()
	}
	financialYearOneStart := time.Date(financialYearOneStartYear, time.April, 1, 0, 0, 0, 0, time.UTC)
	financialYearOneEnd := time.Date(financialYearOneStartYear+1, time.March, 31, 23, 59, 59, 999999999, time.UTC)

	var financialYearTwoStartYear int
	if d1.Time.Month() < time.April {
		financialYearTwoStartYear = d1.Time.Year() - 1
	} else {
		financialYearTwoStartYear = d1.Time.Year()
	}
	financialYearTwoStart := time.Date(financialYearTwoStartYear, time.April, 1, 0, 0, 0, 0, time.UTC)
	financialYearTwoEnd := time.Date(financialYearTwoStartYear+1, time.March, 31, 23, 59, 59, 999999999, time.UTC)

	return financialYearOneStart == financialYearTwoStart && financialYearOneEnd == financialYearTwoEnd
}

func (d Date) IsNull() bool {
	nullDate := NewDate("01/01/0001")
	return d.Time.Equal(nullDate.Time)
}

func (d *Date) UnmarshalJSON(b []byte) error {
	t, err := stringToTime(string(b))
	if err != nil {
		return err
	}

	*d = Date{Time: t}
	return nil
}

func stringToTime(s string) (time.Time, error) {
	value := strings.Trim(string(s), `"`)
	if value == "" || value == "null" {
		return time.Time{}, nil
	}

	value = strings.ReplaceAll(value, `\`, "")
	supportedFormats := []string{
		"02/01/2006",
		"2006-01-02T15:04:05+00:00",
	}

	var t time.Time
	var err error

	for _, format := range supportedFormats {
		t, err = time.Parse(format, value)
		if err != nil {
			continue
		}
		break
	}
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Time.Format("02\\/01\\/2006") + `"`), nil
}
