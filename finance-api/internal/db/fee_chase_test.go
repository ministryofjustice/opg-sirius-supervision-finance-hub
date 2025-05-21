package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_fee_chase() {
	ctx := suite.ctx

	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)
	//twoYearsAgo := today.Sub(2, 0, 0)

	// client 1 with an unpaid invoice
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234")
	client1DeputyID := suite.seeder.CreateDeputy(ctx, client1ID, "Barry", "Manilow", "PRO")
	suite.seeder.CreateAddresses(ctx, client1DeputyID, []string{"91 Fake Avenue", "Binglestone View"}, "Realton", "Nonfictionshire", "OK1 NO2", true)

	suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, nil)

	// client 2 with two invoices - one fully paid, one partially paid
	client2ID := suite.seeder.CreateClient(ctx, "Wallace", "Gromit", "87654321", "4321")
	client2DeputyID := suite.seeder.CreateDeputy(ctx, client2ID, "Jeffrey", "Buckley", "PA")
	suite.seeder.CreateAddresses(ctx, client2DeputyID, []string{"92 Fake Avenue", "Binglestone View", "Greater Gregley"}, "Blompton", "Heartwoodshire", "NO2 RLY", false)

	suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, valToPtr("120"), yesterday.StringPtr(), nil, nil, nil, nil)

	c := Client{suite.seeder.Conn}

	rows, err := c.Run(ctx, &FeeChase{})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// client 1
	assert.Equal(suite.T(), "12345678", results[0]["Case_no"], "Case name - client 1")
	assert.Equal(suite.T(), "1", results[0]["Client_no"], "Client number - client 1")
	assert.Equal(suite.T(), "", results[0]["Client_title"], "Client title - client 1")
	assert.Equal(suite.T(), "Ian", results[0]["Client_forename"], "Client forename - client 1")
	assert.Equal(suite.T(), "Test", results[0]["Client_surname"], "Client surname - client 1")
	assert.Equal(suite.T(), "No", results[0]["Do_not_chase"], "Do not chase - client 1")
	assert.Equal(suite.T(), "Demanded", results[0]["Payment_method"], "Payment method - client 1")
	assert.Equal(suite.T(), "PRO", results[0]["Deputy_type"], "Deputy type - client 1")
	assert.Equal(suite.T(), "", results[0]["Deputy_no"], "Deputy number - client 1")
	assert.Equal(suite.T(), "Yes", results[0]["Airmail"], "Airmail required - client 1")
	assert.Equal(suite.T(), "", results[0]["Deputy_title"], "Deputy title - client 1")
	assert.Equal(suite.T(), "No", results[0]["Deputy_Welsh"], "Deputy Welsh - client 1")
	assert.Equal(suite.T(), "No", results[0]["Deputy_Large_Print"], "Deputy large print - client 1")
	assert.Equal(suite.T(), "Barry Manilow", results[0]["Deputy_name"], "Deputy name - client 1")
	assert.Equal(suite.T(), "", results[0]["Email"], "Deputy email - client 1")
	assert.Equal(suite.T(), "91 Fake Avenue", results[0]["Address1"], "Deputy address 1 - client 1")
	assert.Equal(suite.T(), "Binglestone View", results[0]["Address2"], "Deputy address 2 - client 1")
	assert.Equal(suite.T(), "", results[0]["Address3"], "Deputy address 3 - client 1")
	assert.Equal(suite.T(), "Realton", results[0]["City_Town"], "Deputy town - client 1")
	assert.Equal(suite.T(), "Nonfictionshire", results[0]["County"], "Deputy county - client 1")
	assert.Equal(suite.T(), "OK1 NO2", results[0]["Postcode"], "Deputy postcode - client 1")
	assert.Equal(suite.T(), "100.00", results[0]["Total_debt"], "Total debt - client 1")
	assert.Equal(suite.T(), "map[debt:100.00 reference:AD000001/25]", results[0]["Invoice"], "Invoices - client 1")

	// client 2
	assert.Equal(suite.T(), "87654321", results[1]["Case_no"], "Case name - client 2")
	assert.Equal(suite.T(), "3", results[1]["Client_no"], "Client number - client 2")
	assert.Equal(suite.T(), "", results[1]["Client_title"], "Client title - client 2")
	assert.Equal(suite.T(), "Wallace", results[1]["Client_forename"], "Client forename - client 2")
	assert.Equal(suite.T(), "Gromit", results[1]["Client_surname"], "Client surname - client 2")
	assert.Equal(suite.T(), "No", results[1]["Do_not_chase"], "Do not chase - client 2")
	assert.Equal(suite.T(), "Demanded", results[1]["Payment_method"], "Payment method - client 2")
	assert.Equal(suite.T(), "PA", results[1]["Deputy_type"], "Deputy type - client 2")
	assert.Equal(suite.T(), "", results[1]["Deputy_no"], "Deputy number - client 2")
	assert.Equal(suite.T(), "No", results[1]["Airmail"], "Airmail required - client 2")
	assert.Equal(suite.T(), "", results[1]["Deputy_title"], "Deputy title - client 2")
	assert.Equal(suite.T(), "No", results[1]["Deputy_Welsh"], "Deputy Welsh - client 2")
	assert.Equal(suite.T(), "No", results[1]["Deputy_Large_Print"], "Deputy large print - client 2")
	assert.Equal(suite.T(), "Jeffrey Buckley", results[1]["Deputy_name"], "Deputy name - client 2")
	assert.Equal(suite.T(), "", results[1]["Email"], "Deputy email - client 2")
	assert.Equal(suite.T(), "92 Fake Avenue", results[1]["Address1"], "Deputy address 1 - client 2")
	assert.Equal(suite.T(), "Binglestone View", results[1]["Address2"], "Deputy address 2 - client 2")
	assert.Equal(suite.T(), "Greater Gregley", results[1]["Address3"], "Deputy address 3 - client 2")
	assert.Equal(suite.T(), "Blompton", results[1]["City_Town"], "Deputy town - client 2")
	assert.Equal(suite.T(), "Heartwoodshire", results[1]["County"], "Deputy county - client 2")
	assert.Equal(suite.T(), "NO2 RLY", results[1]["Postcode"], "Deputy postcode - client 2")
	assert.Equal(suite.T(), "100.00", results[1]["Total_debt"], "Total debt - client 2")
	assert.Equal(suite.T(), "map[debt:100.00 reference:AD000002/25, debt:1.20 reference:S2000001/25]", results[1]["Invoice"], "Invoices - client 2")
}
