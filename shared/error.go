package shared

type BadRequest struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

func (b BadRequest) Error() string {
	return b.Reason
}
