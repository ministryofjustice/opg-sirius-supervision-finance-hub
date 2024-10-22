package api

import (
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

func (c *ApiClient) GetUser(ctx Context, userId int) (shared.Assignee, error) {
	logger := telemetry.LoggerFromContext(ctx.Context)

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
		logger.Info(fmt.Sprintf("user %d not found - placeholder added to cache", userId))
		placeholder := shared.Assignee{
			Id:          userId,
			DisplayName: "Unknown User",
			Roles:       nil,
		}
		c.caches.updateUsers([]shared.Assignee{placeholder})
		return placeholder, nil
	}

	return *user, err
}
