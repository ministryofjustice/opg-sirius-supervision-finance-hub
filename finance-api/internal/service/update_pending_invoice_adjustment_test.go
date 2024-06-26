package service

import (
	"context"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_UpdatePendingInvoiceAdjustment() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (15, 15, '1234', 'DEMANDED', NULL);",
		"INSERT INTO invoice VALUES (15, 15, 15, 'S2', 'S203531/19', '2019-04-01', '2020-03-31', 12300, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, 12300, '2019-06-06', NULL);",
		"INSERT INTO ledger VALUES (15, 'random1223', '2022-04-11T08:36:40+00:00', '', 12300, '', 'CREDIT MEMO', 'PENDING', 15, 15, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);",
		"INSERT INTO ledger_allocation VALUES (15, 15, 15, '2022-04-11T08:36:40+00:00', 12300, 'PENDING', NULL, 'Notes here', '2022-04-11', NULL);",
	)

	ctx := context.Background()
	Store := store.New(conn)

	s := &Service{
		store: Store,
		tx:    conn,
	}

	err := s.UpdatePendingInvoiceAdjustment(15, "APPROVED")
	if err != nil {
		suite.T().Error("update pending invoice failed")
	}
	row := conn.QueryRow(ctx, "SELECT status FROM supervision_finance.ledger_allocation WHERE id = 15")

	var status string
	_ = row.Scan(&status)

	assert.Equal(suite.T(), "APPROVED", status)
}
