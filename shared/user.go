package shared

type User struct {
	ID          int32    `json:"id"`
	DisplayName string   `json:"displayName"`
	Roles       []string `json:"roles"`
}

func (u User) IsFinanceUser() bool {
	return contains(u.Roles, "Finance User")
}

func (u User) IsFinanceAdmin() bool {
	return contains(u.Roles, "Finance Admin")
}

func (u User) IsFinanceManager() bool {
	return contains(u.Roles, "Finance Manager")
}

func (u User) IsCorporateFinance() bool {
	return contains(u.Roles, "Corporate Finance")
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
