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
	users *cache.Cache
}

func newCaches() *Caches {
	users := cache.New(defaultExpiration, defaultExpiration)
	placeholder := shared.User{
		ID:          0,
		DisplayName: "Unknown User",
		Roles:       nil,
	}
	_ = users.Add("0", &placeholder, cache.NoExpiration)
	return &Caches{
		users: users,
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
