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
	return fmt.Sprintf("bad request: field=%s reason=%s", b.Field, b.Reason)
}

func (b BadRequest) HTTPStatus() int { return http.StatusBadRequest }

func (b BadRequest) HasData() bool {
	return true
}

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

func (b BadRequests) HasData() bool {
	return true
}

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

type ValidationErrors map[string]map[string]string

type ValidationError struct {
	Errors ValidationErrors `json:"validation_errors"`
}

func (ve ValidationError) Error() string {
	return "validation failed"
}

func (ve ValidationError) HTTPStatus() int { return http.StatusUnprocessableEntity }

func (ve ValidationError) HasData() bool {
	return true
}

type Unauthorized struct {
	error
}

func (u Unauthorized) Error() string {
	return "unauthorized"
}

func (u Unauthorized) Unwrap() error {
	return u.error
}

func (u Unauthorized) HTTPStatus() int { return http.StatusUnauthorized }

type InternalServer struct {
	error
}

func InternalServerError(error error) *InternalServer {
	return &InternalServer{error: error}
}

func (s InternalServer) Unwrap() error {
	return s.error
}

func (s InternalServer) Error() string {
	return "server error"
}

func (s InternalServer) HTTPStatus() int { return http.StatusInternalServerError }

type Forbidden struct {
	error
}

func ForbiddenError(error error) *Forbidden {
	return &Forbidden{error: error}
}

func (s Forbidden) Unwrap() error {
	return s.error
}

func (s Forbidden) Error() string {
	return "forbidden"
}

func (s Forbidden) HTTPStatus() int { return http.StatusForbidden }

type StatusError struct {
	StatusCode int
}

func (s StatusError) Error() string {
	return fmt.Sprintf("status error: %d", s)
}

func (s StatusError) HTTPStatus() int {
	return s.StatusCode
}
