package shared

import (
	"fmt"
	"strings"
)

type BadRequest struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

func (b BadRequest) Error() string {
	return b.Reason
}

type BadRequests struct {
	Reasons []string `json:"reasons"`
}

func (b BadRequests) Error() string {
	return fmt.Sprintf("bad requests: %s", strings.Join(b.Reasons, ", "))
}
