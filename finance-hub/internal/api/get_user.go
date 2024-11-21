package api

import (
	"encoding/json"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

func (c *ApiClient) GetUser(ctx Context, userId int) (shared.Assignee, error) {
	user, ok := c.caches.getUser(userId)

	if ok {
		return *user, nil
	}

	req, err := c.newSiriusRequest(ctx, http.MethodGet, "/users", nil)
	if err != nil {
		return shared.Assignee{}, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return shared.Assignee{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return shared.Assignee{}, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return shared.Assignee{}, newStatusError(resp)
	}

	var users []shared.Assignee
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		return shared.Assignee{}, err
	}

	c.caches.updateUsers(users)
	user, ok = c.caches.getUser(userId)
	if !ok {
		return *c.caches.getAndSetPlaceholder(userId), nil
	}

	return *user, err
}
