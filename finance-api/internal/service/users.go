package service

import (
	"context"
	"encoding/json"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
)

func (s *Service) GetCurrentUser() (*shared.Assignee, error) {
	ctx := context.Background()

	user, err := s.Store.CurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	// Parse the JSON data into a map[string]interface{}
	roles, assignee, err2 := s.mapRoles(err, user)
	if err2 != nil {
		return assignee, err2
	}

	return &shared.Assignee{
		Id:    int(user.ID),
		Roles: roles,
	}, nil
}

func (s *Service) mapRoles(err error, user store.User) ([]string, *shared.Assignee, error) {
	var rolesJson map[string]interface{}
	err = json.Unmarshal([]byte(user.Roles), &rolesJson)
	if err != nil {
		return nil, nil, err
	}

	var roles []string
	for k := range rolesJson {
		roles = append(roles, k)
	}
	return roles, nil, nil
}
