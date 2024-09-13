package apierror

import "net/http"

type BadRequest struct {
	error
	Reason string
}

func BadRequestError(error error, reason string) *BadRequest {
	return &BadRequest{error: error, Reason: reason}
}

func (b BadRequest) Error() string {
	return b.Reason
}

func (b BadRequest) HTTPStatus() int { return http.StatusBadRequest }

type NotFound struct {
	error
}

func NotFoundError(error error) *NotFound {
	return &NotFound{error: error}
}

func (n NotFound) Unwrap() error {
	return n.error
}

func (n NotFound) Error() string {
	return "requested resource not found"
}

func (n NotFound) HTTPStatus() int { return http.StatusNotFound }
