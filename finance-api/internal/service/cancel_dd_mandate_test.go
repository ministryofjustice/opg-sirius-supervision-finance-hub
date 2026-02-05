package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_CancelDirectDebitMandate() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	today := time.Now()

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL, '1234567T');",
		fmt.Sprintf("INSERT INTO pending_collection VALUES (1, 1, '%s', 12300, 'PENDING', NULL, '2025-10-10', 1)", today.AddDate(0, 0, 1).Format("2006-01-02")),  // don't cancel but use for closure date calc
		fmt.Sprintf("INSERT INTO pending_collection VALUES (2, 1, '%s', 12300, 'PENDING', NULL, '2025-10-10', 1)", today.AddDate(0, 0, 10).Format("2006-01-02")), // cancel
	)

	Store := store.New(seeder.Conn)
	allpayMock := mockAllpay{}
	dispatchMock := mockDispatch{}
	govUKMock := &mockGovUK{}

	s := &Service{
		store:    Store,
		allpay:   &allpayMock,
		dispatch: &dispatchMock,
		govUK:    govUKMock,
		tx:       seeder.Conn,
	}

	err := s.CancelDirectDebitMandate(ctx, 11, shared.CancelMandate{
		AllPayCustomer: shared.AllPayCustomer{
			ClientReference: "1234567T",
			Surname:         "Nameson",
		},
	})
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "CancelMandate", allpayMock.called[0])
	assert.Equal(suite.T(), today.AddDate(0, 0, 2).UTC().Truncate(24*time.Hour), allpayMock.closureDate)

	rows := seeder.QueryRow(ctx, "SELECT payment_method FROM supervision_finance.finance_client WHERE id = 1")
	var paymentMethod string
	_ = rows.Scan(&paymentMethod)

	assert.Equal(suite.T(), "DEMANDED", paymentMethod)

	rows = seeder.QueryRow(ctx, "SELECT id FROM supervision_finance.pending_collection WHERE status = 'CANCELLED'")
	var id int
	_ = rows.Scan(&id)

	assert.Equal(suite.T(), 2, id)

	assert.Equal(suite.T(), event.PaymentMethod{
		ClientID:      11,
		PaymentMethod: shared.PaymentMethodDemanded,
	}, dispatchMock.event)
}

func (suite *IntegrationSuite) TestService_CancelDirectDebitMandate_fails() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL, '1234567T');",
	)

	Store := store.New(seeder.Conn)
	allpayMock := mockAllpay{
		errs: map[string]error{"CancelMandate": errors.New("some error")},
	}
	dispatchMock := mockDispatch{}
	govUKMock := &mockGovUK{}

	s := &Service{
		store:    Store,
		allpay:   &allpayMock,
		dispatch: &dispatchMock,
		govUK:    govUKMock,
		tx:       seeder.Conn,
	}

	err := s.CancelDirectDebitMandate(ctx, 11, shared.CancelMandate{
		AllPayCustomer: shared.AllPayCustomer{
			ClientReference: "1234567T",
			Surname:         "Nameson",
		},
	})
	assert.Error(suite.T(), err)

	assert.Equal(suite.T(), "CancelMandate", allpayMock.called[0])

	rows := seeder.QueryRow(ctx, "SELECT payment_method FROM supervision_finance.finance_client WHERE id = 1")
	var paymentMethod string
	_ = rows.Scan(&paymentMethod)

	assert.Equal(suite.T(), "DIRECT DEBIT", paymentMethod) // not changed when unable to update in allpay
	assert.Nil(suite.T(), dispatchMock.event)              // event should not have been sent
}

func TestCalculateClosureDate(t *testing.T) {
	govUKMock := mockGovUK{
		NonWorkingDays: []time.Time{
			time.Now().UTC().Truncate(24 * time.Hour),                  // today is a non-working day
			time.Now().UTC().AddDate(0, 0, 4).Truncate(24 * time.Hour), // non-working day on day 4
		},
	}
	s := &Service{
		govUK: &govUKMock,
	}

	tests := []struct {
		name        string
		collections []store.GetPendingCollectionsRow
		want        time.Time
	}{
		{
			name:        "no pending collections",
			collections: []store.GetPendingCollectionsRow{},
			want:        time.Now().AddDate(0, 0, 1), // today is non-working day
		},
		{
			name: "pending collection on day 1",
			collections: []store.GetPendingCollectionsRow{
				{
					ID:     1,
					Amount: 12345,
					CollectionDate: pgtype.Date{
						Time:             time.Now().AddDate(0, 0, 1),
						InfinityModifier: 0,
						Valid:            true,
					},
				},
			},
			want: time.Now().AddDate(0, 0, 2), // 1 day after pending collection date
		},
		{
			name: "pending collection on day 2",
			collections: []store.GetPendingCollectionsRow{
				{
					ID:     1,
					Amount: 12345,
					CollectionDate: pgtype.Date{
						Time:             time.Now().AddDate(0, 0, 2),
						InfinityModifier: 0,
						Valid:            true,
					},
				},
			},
			want: time.Now().AddDate(0, 0, 3),
		},
		{
			name: "pending collection on day 3",
			collections: []store.GetPendingCollectionsRow{
				{
					ID:     1,
					Amount: 12345,
					CollectionDate: pgtype.Date{
						Time:             time.Now().AddDate(0, 0, 3),
						InfinityModifier: 0,
						Valid:            true,
					},
				},
			},
			want: time.Now().AddDate(0, 0, 5), // non-working day on day 4
		},
		{
			name: "pending collection on day 4",
			collections: []store.GetPendingCollectionsRow{
				{
					ID:     1,
					Amount: 12345,
					CollectionDate: pgtype.Date{
						Time:             time.Now().AddDate(0, 0, 4),
						InfinityModifier: 0,
						Valid:            true,
					},
				},
			},
			want: time.Now().AddDate(0, 0, 1), // today is non-working day
		},
		{
			name: "multiple pending collections",
			collections: []store.GetPendingCollectionsRow{
				{
					ID:     1,
					Amount: 12345,
					CollectionDate: pgtype.Date{
						Time:             time.Now().AddDate(0, 0, 1),
						InfinityModifier: 0,
						Valid:            true,
					},
				},
				{
					ID:     2,
					Amount: 12345,
					CollectionDate: pgtype.Date{
						Time:             time.Now().AddDate(0, 0, 2),
						InfinityModifier: 0,
						Valid:            true,
					},
				},
				{
					ID:     3,
					Amount: 12345,
					CollectionDate: pgtype.Date{
						Time:             time.Now().AddDate(0, 0, 20),
						InfinityModifier: 0,
						Valid:            true,
					},
				},
			},
			want: time.Now().AddDate(0, 0, 3),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, _ := s.calculateClosureDate(context.Background(), tt.collections)
			assert.Equalf(t, tt.want.UTC().Truncate(24*time.Hour), actual.UTC().Truncate(24*time.Hour), tt.name)
		})
	}
}
