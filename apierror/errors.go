package apierror

import (
	"fmt"
	"net/http"
	"strings"
)

type BadRequest struct {
	error
	Field  string `json:"field,omitempty"`
	Reason string `json:"reason"`
}

func BadRequestError(field string, reason string, error error) *BadRequest {
	return &BadRequest{error: error, Field: field, Reason: reason}
}

func (b BadRequest) Error() string {
	return b.Reason
}

func (b BadRequest) HTTPStatus() int { return http.StatusBadRequest }

type BadRequests struct {
	Reasons []string `json:"reasons"`
}

func BadRequestsError(reasons []string) *BadRequests {
	return &BadRequests{Reasons: reasons}
}

func (b BadRequests) Error() string {
	return fmt.Sprintf("bad requests: %s", strings.Join(b.Reasons, ", "))
}

func (b BadRequests) HTTPStatus() int { return http.StatusBadRequest }

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
