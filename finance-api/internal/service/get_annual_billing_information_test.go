package service

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_GetAnnualBillingInformation() {
	seeder := suite.cm.Seeder(suite.ctx, suite.T())

	//invoice with skipped status should update skipped count
	seeder.SeedData(
		"INSERT INTO supervision_finance.property (id, key, value) VALUES (1, 'AnnualBillingYear', '2025');",
		"INSERT INTO public.persons VALUES (111, NULL, NULL, 'Scheduleson', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'actor_client', 'ACTIVE');",
		"INSERT INTO public.cases (id, client_id, orderstatus) VALUES (101, 111, 'ACTIVE');",
		"INSERT INTO supervision_finance.finance_client (id, client_id, sop_number, payment_method, batchnumber, court_ref) VALUES (222, 111, '1234', 'DEMANDED', NULL, '1234567T');",
		"INSERT INTO supervision_finance.invoice VALUES (333, 111, 222, 'S2', 'S200123/25', '2025-04-02', '2026-03-31', 5000, NULL, '2025-04-02', NULL, '2025-04-02');",
		"INSERT INTO supervision_finance.invoice_email_status VALUES (1, 333, 'SKIPPED', 'af1');",
	)

	//invoice on active order for non deceased client and with no record in invoice_email_status table should return expected (i.e. unprocessed)
	seeder.SeedData(
		"INSERT INTO public.persons VALUES (666, NULL, NULL, 'Scheduleson', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'actor_client', 'ACTIVE');",
		"INSERT INTO public.cases (id, client_id, orderstatus) VALUES (109, 666, 'ACTIVE');",
		"INSERT INTO supervision_finance.finance_client (id, client_id, sop_number, payment_method, batchnumber, court_ref) VALUES (878, 666, '1234', 'DEMANDED', NULL, '9898567T');",
		"INSERT INTO supervision_finance.invoice VALUES (356, 666, 878, 'S2', 'S255523/25', '2025-04-02', '2026-03-31', 5000, NULL, '2025-04-02', NULL, '2025-04-02');",
	)

	//invoice with processed status should update issued count
	seeder.SeedData(
		"INSERT INTO public.persons VALUES (555, NULL, NULL, 'Scheduleson', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'actor_client', 'ACTIVE');",
		"INSERT INTO public.cases (id, client_id, orderstatus) VALUES (102, 555, 'ACTIVE');",
		"INSERT INTO supervision_finance.finance_client (id, client_id, sop_number, payment_method, batchnumber, court_ref) VALUES (999, 555, '2222', 'DEMANDED', NULL, '3334567T');",
		"INSERT INTO supervision_finance.invoice VALUES (334, 555, 999, 'S2', 'S200125/25', '2025-04-02', '2026-03-31', 5000, NULL, '2025-04-02', NULL, '2025-04-02');",
		"INSERT INTO supervision_finance.invoice_email_status VALUES (2, 334, 'PROCESSED', 'af2');",
	)

	//invoice with in progress status should update issued count
	seeder.SeedData(
		"INSERT INTO public.persons VALUES (556, NULL, NULL, 'Scheduleson', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'actor_client', 'ACTIVE');",
		"INSERT INTO public.cases (id, client_id, orderstatus) VALUES (103, 556, 'ACTIVE');",
		"INSERT INTO supervision_finance.finance_client (id, client_id, sop_number, payment_method, batchnumber, court_ref) VALUES (998, 556, '2323', 'DEMANDED', NULL, '3234567T');",
		"INSERT INTO supervision_finance.invoice VALUES (335, 556, 998, 'S2', 'S211135/25', '2025-04-02', '2026-03-31', 5000, NULL, '2025-04-02', NULL, '2025-04-02');",
		"INSERT INTO supervision_finance.invoice_email_status VALUES (3, 335, 'IN_PROGRESS', 'af2');",
	)

	//invoice with error status should update issued count
	seeder.SeedData(
		"INSERT INTO public.persons VALUES (557, NULL, NULL, 'Scheduleson', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'actor_client', 'ACTIVE');",
		"INSERT INTO public.cases (id, client_id, orderstatus) VALUES (104, 557, 'ACTIVE');",
		"INSERT INTO supervision_finance.finance_client (id, client_id, sop_number, payment_method, batchnumber, court_ref) VALUES (997, 557, '2221', 'DEMANDED', NULL, '1134567T');",
		"INSERT INTO supervision_finance.invoice VALUES (332, 557, 997, 'S2', 'S200225/25', '2025-04-02', '2026-03-31', 5000, NULL, '2025-04-02', NULL, '2025-04-02');",
		"INSERT INTO supervision_finance.invoice_email_status VALUES (4, 332, 'ERROR', 'af2');",
	)

	//invoice with none status should update issued count
	seeder.SeedData(
		"INSERT INTO public.persons VALUES (558, NULL, NULL, 'Scheduleson', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'actor_client', 'ACTIVE');",
		"INSERT INTO public.cases (id, client_id, orderstatus) VALUES (105, 558, 'ACTIVE');",
		"INSERT INTO supervision_finance.finance_client (id, client_id, sop_number, payment_method, batchnumber, court_ref) VALUES (996, 558, '2441', 'DEMANDED', NULL, '3334567T');",
		"INSERT INTO supervision_finance.invoice VALUES (402, 558, 996, 'S2', 'S243225/25', '2025-04-02', '2026-03-31', 5000, NULL, '2025-04-02', NULL, '2025-04-02');",
		"INSERT INTO supervision_finance.invoice_email_status VALUES (5, 402, 'ERROR', 'af2');",
	)

	Store := store.New(seeder)
	s := &Service{
		store: Store,
	}
	got, _ := s.GetAnnualBillingInfo(suite.ctx)

	assert.Equal(suite.T(), int64(1), got.DemandedExpectedCount)
	assert.Equal(suite.T(), int64(4), got.DemandedIssuedCount)
	assert.Equal(suite.T(), int64(1), got.DemandedSkippedCount)
	assert.Equal(suite.T(), "2025", got.AnnualBillingYear)
}
