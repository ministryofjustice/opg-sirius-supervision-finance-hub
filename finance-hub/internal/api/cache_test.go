package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCache_updateHolidays(t *testing.T) {
	caches := newCaches()
	caches.holidays.Set("refresh", false, defaultExpiration)

	caches.updateHolidays([]Holiday{{Date: "2025-01-01"}})

	v, _ := caches.holidays.Get("refresh")
	assert.True(t, v.(bool))

	v, _ = caches.holidays.Get("2025-01-01")
	assert.True(t, v.(bool))
}
