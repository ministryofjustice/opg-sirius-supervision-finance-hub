package shared

const ErrUnauthorized ClientError = "unauthorized"

type ClientError string

func (e ClientError) Error() string {
	return string(e)
}

type ValidationErrors map[string]map[string]string

type ValidationError struct {
	Message string
	Errors  ValidationErrors `json:"validation_errors"`
}

func (ve ValidationError) Error() string {
	return ve.Message
}
