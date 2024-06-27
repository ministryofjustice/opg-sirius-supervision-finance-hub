package api

import (
	"encoding/json"
	"errors"
	"github.com/opg-sirius-finance-hub/shared"
	"log"
	"net/http"
)

func (c *ApiClient) GetUser(ctx Context, userId int) (shared.Assignee, error) {
	user, ok := c.caches.getUser(userId)

	if ok {
		log.Printf("cache hit for user %d", userId)
		return *user, nil
	}

	log.Printf("no cache for user %d, fetching", userId)
	req, err := c.newSiriusRequest(ctx, http.MethodGet, "/api/v1/users", nil)
	if err != nil {
		c.logErrorRequest(req, err)
		return shared.Assignee{}, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		c.logger.Request(req, err)
		return shared.Assignee{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		c.logger.Request(req, err)
		return shared.Assignee{}, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Request(req, err)
		return shared.Assignee{}, newStatusError(resp)
	}

	var users []shared.Assignee
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		c.logger.Request(req, err)
		return shared.Assignee{}, err
	}

	c.caches.updateUsers(users)
	log.Println("cache updated")
	user, ok = c.caches.getUser(userId)
	if !ok {
		return shared.Assignee{}, errors.New("user not found")
	}

	return *user, err
}