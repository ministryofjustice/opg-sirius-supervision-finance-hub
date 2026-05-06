package service

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func TestService_QueueScheduleRemovals(t *testing.T) {
	records := [][]string{
		{"Client Reference", "Surname", "Amount"},
		{"123ABC", "Felicity", "12321"},
		{"321CBA", "Brian", "32123"},
	}
	uploadDate := shared.NewDate("2022-04-02")

	dispatch := &mockDispatch{}
	s := Service{dispatch: dispatch}
	s.QueueScheduleRemovals(context.Background(), records, uploadDate)

	assert.Len(t, dispatch.called, 2)

	lastEvent := dispatch.event.(event.ScheduleToRemove)
	assert.Equal(t, "321CBA", lastEvent.CourtRef)
	assert.Equal(t, "Brian", lastEvent.Surname)
	assert.Equal(t, 32123, lastEvent.Amount)
	assert.Equal(t, uploadDate, lastEvent.Date)
}
