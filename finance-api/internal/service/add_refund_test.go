package service

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"time"
)

func (suite *IntegrationSuite) TestService_AddRefund() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())
	dispatch := &mockDispatch{}
	s := NewService(seeder.Conn, dispatch, nil, nil, nil)

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (24, 2401, '1234', 'DEMANDED', NULL);",
		"INSERT INTO ledger VALUES (1, 'overpayment', '2024-01-02 15:32:10', '', 5000, 'payment 1', 'MOTO CARD PAYMENT', 'CONFIRMED', 24, NULL, NULL, NULL, '2024-01-01', NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, NULL, '2024-01-02 15:32:10', -5000, 'UNAPPLIED', NULL, '', '2024-01-01', NULL);",
	)

	params := shared.AddRefund{
		AccountName:   "Reginald Refund",
		AccountNumber: "12345678",
		SortCode:      "11-22-33",
		RefundNotes:   "A refund note",
	}

	clientID := int32(2401)
	err := s.AddRefund(ctx, clientID, params)
	if err != nil {
		suite.T().Error("Add refund failed")
	}

	rows := seeder.QueryRow(ctx, "SELECT r.raised_date, r.amount, r.decision, r.notes, r.created_by, b.name, b.account, b.sort_code FROM refund r JOIN bank_details b ON r.id = b.refund_id WHERE r.finance_client_id = (SELECT id FROM finance_client WHERE client_id = $1)", clientID)

	var (
		raisedDate  time.Time
		amount      int
		decision    string
		notes       string
		createdById int
		name        string
		account     string
		sortCode    string
	)

	_ = rows.Scan(
		&raisedDate,
		&amount,
		&decision,
		&notes,
		&createdById,
		&name,
		&account,
		&sortCode,
	)

	assert.Equal(suite.T(), time.Now().Format("2006-01-02"), raisedDate.Format("2006-01-02"))
	assert.Equal(suite.T(), 5000, amount)
	assert.Equal(suite.T(), "PENDING", decision)
	assert.Equal(suite.T(), params.RefundNotes, notes)
	assert.Equal(suite.T(), 10, createdById)
	assert.Equal(suite.T(), params.AccountName, name)
	assert.Equal(suite.T(), params.AccountNumber, account)
	assert.Equal(suite.T(), params.SortCode, sortCode)

	assert.Equal(suite.T(), clientID, dispatch.event.(event.RefundAdded).ClientID)
}

func (suite *IntegrationSuite) TestService_AddRefund_noCreditToRefund() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())
	s := NewService(seeder.Conn, nil, nil, nil, nil)

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (24, 24, '1234', 'DEMANDED', NULL);",
		"INSERT INTO ledger VALUES (1, 'overpayment', '2024-01-02 15:32:10', '', 5000, 'payment 1', 'MOTO CARD PAYMENT', 'CONFIRMED', 24, NULL, NULL, NULL, '2024-01-01', NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, NULL, '2024-01-02 15:32:10', -5000, 'UNAPPLIED', NULL, '', '2024-01-01', NULL);",
		"INSERT INTO ledger VALUES (2, 'existing refund', '2024-01-02 15:32:10', '', 5000, 'refund 1', 'MOTO CARD PAYMENT', 'CONFIRMED', 24, NULL, NULL, NULL, '2024-01-01', NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (2, 2, NULL, '2024-01-02 15:32:10', 5000, 'REAPPLIED', NULL, '', '2024-01-01', NULL);",
	)

	params := shared.AddRefund{
		AccountName:   "Reginald Refund",
		AccountNumber: "12345678",
		SortCode:      "11-22-33",
		RefundNotes:   "A refund note",
	}

	clientID := int32(24)

	expectedErr := apierror.BadRequest{Reason: "NoCreditToRefund"}

	err := s.AddRefund(suite.ctx, clientID, params)
	if err != nil {
		assert.Equal(suite.T(), expectedErr, err)
	}
}
