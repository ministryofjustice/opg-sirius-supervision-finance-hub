package shared

import (
	"strings"
)

type Assignee struct {
	Id          int      `json:"id"`
	DisplayName string   `json:"displayName"`
	Roles       []string `json:"roles"`
}

func (m Assignee) GetRoles() string {
	return strings.Join(m.Roles, ",")
}
