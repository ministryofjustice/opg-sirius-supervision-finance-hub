package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_UpdatePendingInvoiceAdjustment(t *testing.T) {
	conn := testDB.GetConn()
	t.Cleanup(func() {
		testDB.Restore()
	})
	conn.SeedData(
		"INSERT INTO finance_client VALUES (15, 15, '1234', 'DEMANDED', null, 12300, 2222);",
		"INSERT INTO invoice VALUES (15, 15, 15, 'S2', 'S203531/19', '2019-04-01', '2020-03-31', 12300, null, '2020-03-20',1, '2020-03-16', 10, null, 12300, '2019-06-06', null);",
		"INSERT INTO ledger VALUES (15, 'random1223', '2022-04-11T08:36:40+00:00', '', 12300, '', 'CREDIT MEMO', 'PENDING', 15, 15, null, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);",
		"INSERT INTO ledger_allocation VALUES (15, 15, 15, '2022-04-11T08:36:40+00:00', 12300, 'PENDING', null, 'Notes here', '2022-04-11', null);",
	)

	ctx := context.Background()
	Store := store.New(conn)

	s := &Service{
		Store: Store,
		TX:    conn,
	}

	err := s.UpdatePendingInvoiceAdjustment(15)
	rows, _ := conn.Query(ctx, "SELECT * FROM supervision_finance.ledger_allocation WHERE id = 15")
	defer rows.Close()

	for rows.Next() {
		var (
			ID              int32
			LedgerID        pgtype.Int4
			InvoiceID       pgtype.Int4
			Datetime        pgtype.Timestamp
			Amount          int32
			Status          string
			Reference       pgtype.Text
			Notes           pgtype.Text
			Allocateddate   pgtype.Date
			Batchnumber     pgtype.Int4
			Source          pgtype.Text
			TransactionType pgtype.Text
		)

		_ = rows.Scan(&ID, &LedgerID, &InvoiceID, &Datetime, &Amount, &Status, &Reference, &Notes, &Allocateddate, &Batchnumber, &Source, &TransactionType)

		assert.Equal(t, "APPROVED", Status)
	}

	if err == nil {
		return
	}
	t.Error("update pending invoice failed")
}