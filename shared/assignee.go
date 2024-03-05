package shared

import (
	"strconv"
	"strings"
)

type Assignee struct {
	Id    int      `json:"id"`
	Roles []string `json:"roles"`
}

func (m Assignee) IsSelected(selectedAssignees []string) bool {
	for _, a := range selectedAssignees {
		id, _ := strconv.Atoi(a)
		if m.Id == id {
			return true
		}
	}
	return false
}

func (m Assignee) GetRoles() string {
	return strings.Join(m.Roles, ",")
}
