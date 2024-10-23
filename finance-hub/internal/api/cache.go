package api

import (
	"github.com/opg-sirius-finance-hub/shared"
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
	placeholder := shared.Assignee{
		Id:          0,
		DisplayName: "Unknown User",
		Roles:       nil,
	}
	_ = users.Add("0", &placeholder, cache.NoExpiration)
	return &Caches{
		users: users,
	}
}

func (c Caches) getUser(id int) (*shared.Assignee, bool) {
	get, b := c.users.Get(strconv.Itoa(id))
	if b {
		return get.(*shared.Assignee), true
	} else {
		return nil, false
	}
}

// getAndSetPlaceholder gets the placeholder user and adds it for the id. This prevents subsequent cache requests for the
// same value forcing a cache refresh.
func (c Caches) getAndSetPlaceholder(id int) *shared.Assignee {
	u, _ := c.users.Get("0")
	placeholder := u.(*shared.Assignee)
	_ = c.users.Add(strconv.Itoa(id), &placeholder, defaultExpiration)
	return u.(*shared.Assignee)
}

func (c Caches) updateUsers(users []shared.Assignee) {
	for _, user := range users {
		_ = c.users.Add(strconv.Itoa(user.Id), &user, defaultExpiration)
	}
}
