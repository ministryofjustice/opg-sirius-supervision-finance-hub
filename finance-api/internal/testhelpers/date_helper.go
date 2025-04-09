package testhelpers

import (
	"fmt"
	"time"
)

type DateHelper struct {
	date time.Time
}

func (s *Seeder) Today() DateHelper {
	date := DateHelper{date: time.Now()}
	return date
}

func (t DateHelper) Add(years int, months int, days int) DateHelper {
	t.date = t.date.AddDate(years, months, days)
	return t
}

func (t DateHelper) Sub(years int, months int, days int) DateHelper {
	t.date = t.date.AddDate(-years, -months, -days)
	return t
}

func (t DateHelper) String() string {
	return t.date.Format("2006-01-02")
}

func (t DateHelper) StringPtr() *string {
	s := t.date.Format("2006-01-02")
	return &s
}

func (t DateHelper) UKString() string {
	return t.date.Format("02/01/2006")
}

func (t DateHelper) Date() time.Time {
	return t.date
}

func (t DateHelper) DatePtr() *time.Time {
	d := t.date
	return &d
}

func (t DateHelper) FinancialYear() string {
	if t.date.Month() >= time.April {
		return fmt.Sprintf("%d/%s", t.date.Year(), t.date.AddDate(1, 0, 0).Format("06"))
	}
	return fmt.Sprintf("%d/%s", t.date.Year()-1, t.date.Format("06"))
}
