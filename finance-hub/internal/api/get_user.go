package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (c *Client) GetUser(ctx context.Context, userId int) (shared.User, error) {
	user, ok := c.caches.getUser(userId)

	if ok {
		return *user, nil
	}

	req, err := c.newSiriusRequest(ctx, http.MethodGet, "/users", nil)
	if err != nil {
		return shared.User{}, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return shared.User{}, err
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode == http.StatusUnauthorized {
		return shared.User{}, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return shared.User{}, newStatusError(resp)
	}

	var users []shared.User
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		return shared.User{}, err
	}

	c.caches.updateUsers(users)
	user, ok = c.caches.getUser(userId)
	if !ok {
		return *c.caches.getAndSetPlaceholder(userId), nil
	}

	return *user, err
}
