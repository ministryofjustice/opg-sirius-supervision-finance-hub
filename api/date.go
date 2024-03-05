package api

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
