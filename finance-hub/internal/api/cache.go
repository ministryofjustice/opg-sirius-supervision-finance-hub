package api

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/patrickmn/go-cache"
	"strconv"
	"time"
)

const (
	defaultExpiration = 12 * time.Hour
)

type Caches struct {
	users    *cache.Cache
	holidays *cache.Cache
}

func newCaches() *Caches {
	users := cache.New(defaultExpiration, defaultExpiration)
	placeholder := shared.User{
		ID:          0,
		DisplayName: "Unknown User",
		Roles:       nil,
	}
	_ = users.Add("0", &placeholder, cache.NoExpiration)

	holidays := cache.New(defaultExpiration, defaultExpiration)
	// unreachable cached value used for triggering cache refresh
	_ = holidays.Add("refresh", true, defaultExpiration)
	return &Caches{
		users:    users,
		holidays: holidays,
	}
}

func (c Caches) getUser(id int) (*shared.User, bool) {
	get, b := c.users.Get(strconv.Itoa(id))
	if b {
		return get.(*shared.User), true
	} else {
		return nil, false
	}
}

// getAndSetPlaceholder gets the placeholder user and adds it for the id. This prevents subsequent cache requests for the
// same value forcing a cache refresh.
func (c Caches) getAndSetPlaceholder(id int) *shared.User {
	u, _ := c.users.Get("0")
	placeholder := u.(*shared.User)
	_ = c.users.Add(strconv.Itoa(id), placeholder, defaultExpiration)
	return u.(*shared.User)
}

func (c Caches) updateUsers(users []shared.User) {
	for _, user := range users {
		_ = c.users.Add(strconv.Itoa(int(user.ID)), &user, defaultExpiration)
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
