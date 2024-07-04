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
	Cache := cache.New(defaultExpiration, defaultExpiration)
	return &Caches{
		users: Cache,
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

func (c Caches) updateUsers(users []shared.Assignee) {
	for _, user := range users {
		_ = c.users.Add(strconv.Itoa(user.Id), &user, defaultExpiration)
	}
}
