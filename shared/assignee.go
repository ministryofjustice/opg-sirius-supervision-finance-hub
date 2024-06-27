package shared

import (
	"strings"
)

type Assignee struct {
	Id      int      `json:"id"`
	Name    string   `json:"name"`
	Surname string   `json:"surname"`
	Roles   []string `json:"roles"`
}

func (m Assignee) GetRoles() string {
	return strings.Join(m.Roles, ",")
}
