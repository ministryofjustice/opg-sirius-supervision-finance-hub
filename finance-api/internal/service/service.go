package service

import (
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
)

type Service struct {
	Store *store.Queries
}

func normaliseNilArrays[T any](arr []T) []T {
	if arr == nil {
		arr = []T{}
	}
	return arr
}
