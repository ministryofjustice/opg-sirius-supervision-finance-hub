package shared

type User struct {
	ID          int32    `json:"id"`
	DisplayName string   `json:"displayName"`
	Roles       []string `json:"roles"`
}
