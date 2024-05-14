package shared

type ValidationErrors map[string]map[string]string

type ValidationError struct {
	Message string
	Errors  ValidationErrors `json:"validation_errors"`
}

func (ve ValidationError) Error() string {
	return ve.Message
}
