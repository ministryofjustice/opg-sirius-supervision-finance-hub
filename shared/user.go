package shared

const (
	RoleFinanceUser      = "Finance User"
	RoleFinanceManager   = "Finance Manager"
	RoleFinanceReporting = "Finance Reporting"
	RoleCorporateFinance = "Corporate Finance"
	RoleAny              = ""
)

type User struct {
	ID          int      `json:"id"`
	DisplayName string   `json:"displayName"`
	Roles       []string `json:"roles"`
}

func (u User) IsFinanceUser() bool {
	return contains(u.Roles, RoleFinanceUser)
}

func (u User) IsFinanceManager() bool {
	return contains(u.Roles, RoleFinanceManager)
}

func (u User) IsFinanceReporting() bool {
	return contains(u.Roles, RoleFinanceReporting)
}

func (u User) IsCorporateFinance() bool {
	return contains(u.Roles, RoleCorporateFinance)
}

func (u User) HasRole(role string) bool {
	if role == RoleAny {
		return true
	}
	return contains(u.Roles, role)
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
