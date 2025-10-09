package shared

import (
	"strconv"
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

func (d Date) IsSameFinancialYear(d1 Date) bool {
	financialYearOneStartYear := d.Time.Year()
	if d.Time.Month() < time.April {
		financialYearOneStartYear = d.Time.Year() - 1
	}

	financialYearTwoStartYear := d1.Time.Year()
	if d1.Time.Month() < time.April {
		financialYearTwoStartYear = d1.Time.Year() - 1
	}

	return financialYearOneStartYear == financialYearTwoStartYear
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
	value := strings.Trim(s, `"`)
	if value == "" || value == "null" {
		return time.Time{}, nil
	}

	value = strings.ReplaceAll(value, `\`, "")
	supportedFormats := []string{
		"02/01/2006",
		"2006-01-02T15:04:05+00:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
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

func (d Date) CalculateFinanceYear() string {
	financialYearOneStartYear := d.Time.Year()
	if d.Time.Month() >= time.January && d.Time.Month() < time.April {
		financialYearOneStartYear = d.Time.Year() - 1
	}

	return strconv.Itoa(financialYearOneStartYear % 100)
}
