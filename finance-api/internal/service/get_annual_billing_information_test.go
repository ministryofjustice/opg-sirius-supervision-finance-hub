package service

import (
	"fmt"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_GetAnnualBillingInformation() {
	seeder := suite.cm.Seeder(suite.ctx, suite.T())
	suite.dataSeedingForYear(seeder, "2025")

	//invoice with skipped status should update skipped count
	suite.dataSeedingForGetBillingInformation(seeder, "ACTIVE", "DEMANDED", "SKIPPED", 0, "S2", "2025-04-02", "2026-03-31")

	//invoice on active order for non deceased client and with no record in invoice_email_status table should return expected (i.e. unprocessed)
	suite.dataSeedingForGetBillingInformation(seeder, "ACTIVE", "DEMANDED", "", 1, "S2", "2025-04-02", "2026-03-31")

	//invoice with processed status should update issued count
	suite.dataSeedingForGetBillingInformation(seeder, "ACTIVE", "DEMANDED", "PROCESSED", 2, "S2", "2025-04-02", "2026-03-31")

	//invoice with in progress status should update issued count
	suite.dataSeedingForGetBillingInformation(seeder, "ACTIVE", "DEMANDED", "IN_PROGRESS", 3, "S2", "2025-04-02", "2026-03-31")

	//invoice with error status should update issued count
	suite.dataSeedingForGetBillingInformation(seeder, "ACTIVE", "DEMANDED", "ERROR", 4, "S2", "2025-04-02", "2026-03-31")

	//invoice with none status should update issued count
	suite.dataSeedingForGetBillingInformation(seeder, "ACTIVE", "DEMANDED", "NONE", 5, "S2", "2025-04-02", "2026-03-31")

	Store := store.New(seeder)
	s := &Service{
		store: Store,
	}
	got, _ := s.GetAnnualBillingInformation(suite.ctx)

	assert.Equal(suite.T(), 1, got.DemandedExpectedCount)
	assert.Equal(suite.T(), 4, got.DemandedIssuedCount)
	assert.Equal(suite.T(), 1, got.DemandedSkippedCount)
	assert.Equal(suite.T(), 2025, got.AnnualBillingYear)
}

func (suite *IntegrationSuite) TestService_GetAnnualBillingInformationForDirectDebitClients() {
	seeder := suite.cm.Seeder(suite.ctx, suite.T())
	suite.dataSeedingForYear(seeder, "2025")

	//invoice on active order for non deceased client and with no record in invoice_email_status table should return expected (i.e. unprocessed)
	suite.dataSeedingForGetBillingInformation(seeder, "ACTIVE", "DIRECT DEBIT", "", 0, "S2", "2025-04-02", "2026-03-31")

	//invoice with processed status should update issued count
	suite.dataSeedingForGetBillingInformation(seeder, "ACTIVE", "DIRECT DEBIT", "PROCESSED", 1, "S2", "2025-04-02", "2026-03-31")

	//invoice with in progress status should update issued count
	suite.dataSeedingForGetBillingInformation(seeder, "ACTIVE", "DIRECT DEBIT", "IN_PROGRESS", 2, "S2", "2025-04-02", "2026-03-31")

	//invoice with error status should update issued count
	suite.dataSeedingForGetBillingInformation(seeder, "ACTIVE", "DIRECT DEBIT", "ERROR", 3, "S2", "2025-04-02", "2026-03-31")

	//invoice with none status should update issued count
	suite.dataSeedingForGetBillingInformation(seeder, "ACTIVE", "DIRECT DEBIT", "ERROR", 4, "S2", "2025-04-02", "2026-03-31")

	Store := store.New(seeder)
	s := &Service{
		store: Store,
	}
	got, _ := s.GetAnnualBillingInformation(suite.ctx)

	assert.Equal(suite.T(), 4, got.DirectDebitIssuedCount)
	assert.Equal(suite.T(), 1, got.DirectDebitExpectedCount)
	assert.Equal(suite.T(), 2025, got.AnnualBillingYear)
}

func (suite *IntegrationSuite) TestService_GetAnnualBillingInformationWillNotCountInvalidCases() {
	seeder := suite.cm.Seeder(suite.ctx, suite.T())
	suite.dataSeedingForYear(seeder, "2025")

	//will ignore cases which are not active - closed
	suite.dataSeedingForGetBillingInformation(seeder, "CLOSED", "DIRECT DEBIT", "", 0, "S2", "2025-04-02", "2026-03-31")

	//will ignore cases which are not active - open
	suite.dataSeedingForGetBillingInformation(seeder, "OPEN", "DEMANDED", "", 1, "S2", "2025-04-02", "2026-03-31")

	//will ignore cases which are not active - duplicate
	suite.dataSeedingForGetBillingInformation(seeder, "DUPLICATE", "DIRECT DEBIT", "", 2, "S2", "2025-04-02", "2026-03-31")

	//will ignore death notified clients
	suite.dataSeedingForGetBillingInformation(seeder, "DEATH_NOTIFIED", "DIRECT DEBIT", "", 3, "S2", "2025-04-02", "2026-03-31")

	//will ignore N2 type fee
	suite.dataSeedingForGetBillingInformation(seeder, "ACTIVE", "DEMANDED", "", 4, "N2", "2025-04-02", "2026-03-31")

	//will ignore N3 type fee
	suite.dataSeedingForGetBillingInformation(seeder, "ACTIVE", "DIRECT DEBIT", "", 5, "N3", "2025-04-02", "2026-03-31")

	//will ignore invoice from before period starts
	suite.dataSeedingForGetBillingInformation(seeder, "ACTIVE", "DEMANDED", "", 6, "S2", "2024-04-02", "2025-03-31")

	//will ignore invoice from after end of period
	suite.dataSeedingForGetBillingInformation(seeder, "ACTIVE", "DIRECT DEBIT", "", 7, "S2", "2026-04-02", "2027-03-31")

	Store := store.New(seeder)
	s := &Service{
		store: Store,
	}
	got, _ := s.GetAnnualBillingInformation(suite.ctx)

	assert.Equal(suite.T(), 0, got.DirectDebitExpectedCount)
	assert.Equal(suite.T(), 0, got.DemandedExpectedCount)
	assert.Equal(suite.T(), 0, got.DirectDebitIssuedCount)
	assert.Equal(suite.T(), 0, got.DemandedIssuedCount)
	assert.Equal(suite.T(), 0, got.DemandedSkippedCount)
	assert.Equal(suite.T(), 2025, got.AnnualBillingYear)
}

func (suite *IntegrationSuite) dataSeedingForYear(
	seeder *testhelpers.Seeder,
	annualBillingYear string,
) {
	seeder.SeedData(
		fmt.Sprintf("INSERT INTO supervision_finance.property (id, key, value) VALUES (1, 'AnnualBillingYear', %s);", annualBillingYear),
	)
}

func (suite *IntegrationSuite) dataSeedingForGetBillingInformation(
	seeder *testhelpers.Seeder,
	orderStatus string,
	paymentMethod string,
	invoiceEmailStatus string,
	uniqueClientAddition int,
	invoiceFeeType string,
	invoiceStartDate string,
	invoiceEndDate string,
) {
	clientId := 665 + uniqueClientAddition
	caseId := 100 + uniqueClientAddition
	fcId := 50 + uniqueClientAddition
	invoiceId := 20 + uniqueClientAddition
	invoiceEmailId := 1 + uniqueClientAddition
	invoiceRef := "S" + fmt.Sprintf("%d", uniqueClientAddition) + "1111/25"
	courtRef := fmt.Sprintf("%d", uniqueClientAddition) + "998567T"

	seeder.SeedData(
		fmt.Sprintf("INSERT INTO public.persons (id, salutation, firstname, surname, caserecnumber, feepayer_id, deputytype, deputynumber, correspondencebywelsh, specialcorrespondencerequirements_largeprint, organisationname, email, type, clientstatus) VALUES (%d, NULL, NULL, 'Scheduleson', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'actor_client', 'ACTIVE');", clientId),
		fmt.Sprintf("INSERT INTO public.cases (id, client_id, orderstatus) VALUES (%d, %d, '%s');", caseId, clientId, orderStatus),
		fmt.Sprintf("INSERT INTO supervision_finance.finance_client (id, client_id, sop_number, payment_method, batchnumber, court_ref) VALUES (%d, %d, '1234', '%s', NULL, '%s');", fcId, clientId, paymentMethod, courtRef),
		fmt.Sprintf("INSERT INTO supervision_finance.invoice VALUES (%d, %d, %d, '%s', '%s', '%s', '%s', 5000, NULL, '2025-04-02', NULL, '2025-04-02');", invoiceId, clientId, fcId, invoiceFeeType, invoiceRef, invoiceStartDate, invoiceEndDate),
	)
	if invoiceEmailStatus != "" {
		seeder.SeedData(
			fmt.Sprintf("INSERT INTO supervision_finance.invoice_email_status VALUES (%d, %d, '%s', 'af2');", invoiceEmailId, invoiceId, invoiceEmailStatus),
		)
	}
}
