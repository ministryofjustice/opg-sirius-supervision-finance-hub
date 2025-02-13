package shared

type User struct {
	ID          int      `json:"id"`
	DisplayName string   `json:"displayName"`
	Roles       []string `json:"roles"`
}
