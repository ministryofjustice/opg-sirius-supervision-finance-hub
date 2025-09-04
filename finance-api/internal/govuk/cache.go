package govuk

import (
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	defaultExpiration = 12 * time.Hour
)

type Caches struct {
	holidays *cache.Cache
}

func newCaches() *Caches {
	holidays := cache.New(defaultExpiration, defaultExpiration)
	// unreachable cached value used for triggering cache refresh
	_ = holidays.Add("refresh", true, defaultExpiration)
	return &Caches{
		holidays: holidays,
	}
}

func (c Caches) updateHolidays(holidays []Holiday) {
	for _, holiday := range holidays {
		_ = c.holidays.Add(holiday.Date, true, defaultExpiration)
	}
}

func (c Caches) isHoliday(d time.Time) bool {
	_, b := c.holidays.Get(d.Format("2006-01-02"))
	return b
}

// shouldRefreshHolidays returns true if the default value has expired
func (c Caches) shouldRefreshHolidays() bool {
	_, b := c.holidays.Get("refresh")
	return !b
}
