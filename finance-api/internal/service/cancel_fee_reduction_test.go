package service

import (
	"context"
	"database/sql"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestService_CancelFeeReduction(t *testing.T) {
	conn := testDB.GetConn()
	t.Cleanup(func() {
		testDB.Restore()
	})
	conn.SeedData(
		"INSERT INTO finance_client VALUES (33, 33, '1234', 'DEMANDED', null, 12300, 2222);",
		"INSERT INTO fee_reduction VALUES (33, 33, 'REMISSION', null, '2019-04-01', '2021-03-31', 'Remission to see the notes', false, '2019-05-01');",
	)

	ctx := context.Background()
	Store := store.New(conn)

	s := &Service{
		store: Store,
		tx:    conn,
	}

	err := s.CancelFeeReduction(33)
	rows, _ := conn.Query(ctx, "SELECT * FROM supervision_finance.fee_reduction WHERE id = 33")
	defer rows.Close()

	for rows.Next() {
		var (
			id            int
			financeClient int
			feeType       string
			evidenceType  sql.NullString
			startDate     time.Time
			endDate       time.Time
			notes         string
			deleted       bool
			dateReceived  time.Time
		)

		_ = rows.Scan(&id, &financeClient, &feeType, &evidenceType, &startDate, &endDate, &notes, &deleted, &dateReceived)

		assert.Equal(t, true, deleted)
		assert.Equal(t, "Remission to see the notes", notes)
	}

	if err == nil {
		return
	}
	t.Error("Cancel fee reduction failed")
}
