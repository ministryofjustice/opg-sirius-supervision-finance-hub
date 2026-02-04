package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_fee_chase() {
	ctx := suite.ctx

	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)

	// client 1 with 2 unpaid invoices and an unrelated warning
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234", "ACTIVE")
	client1DeputyID := suite.seeder.CreateDeputy(ctx, client1ID, "Barry", "Manilow", "PRO")
	suite.seeder.CreateOrder(ctx, client1ID, "pfa")
	orderId := suite.seeder.GetLastOrderIdForClient(ctx, client1ID)
	suite.seeder.UpdateHowDeputyAppointed(ctx, orderId, "JOINT")
	suite.seeder.CreateDeputyOrder(ctx, client1DeputyID, orderId)
	suite.seeder.CreateDeputyImportantInformation(ctx, client1DeputyID, "SCHEDULE")
	suite.seeder.CreateAddresses(ctx, client1DeputyID, []string{"91 Fake Avenue", "Binglestone View"}, "Realton", "Nonfictionshire", "OK1 NO2", true)

	suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeB2, valToPtr("90.99"), yesterday.StringPtr(), nil, nil, nil, nil)

	suite.seeder.CreateWarning(ctx, client1ID, "Evil guy")

	// client 2 with two invoices - one fully paid, one partially paid
	client2ID := suite.seeder.CreateClient(ctx, "Wallace", "Gromit", "87654321", "4321", "ACTIVE")
	client2DeputyID := suite.seeder.CreateDeputy(ctx, client2ID, "Jeffrey", "Buckley", "PA")

	suite.seeder.CreateClosedOrder(ctx, client2ID, today.Date(), "")
	orderId = suite.seeder.GetLastOrderIdForClient(ctx, client2ID)
	suite.seeder.UpdateHowDeputyAppointed(ctx, orderId, "SOLE")
	suite.seeder.CreateDeputyOrder(ctx, client2ID, orderId)

	suite.seeder.CreateAddresses(ctx, client2DeputyID, []string{"92 Fake Avenue", "Binglestone View", "Greater Gregley"}, "Blompton", "Heartwoodshire", "NO2 RLY", false)
	suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS2, valToPtr("120.21"), yesterday.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 15000, yesterday.Date(), "87654321", shared.TransactionTypeMotoCardPayment, today.Date(), 0)

	// client 3 with one invoice fully paid
	client3ID := suite.seeder.CreateClient(ctx, "Patrick", "Stewart", "87651234", "4312", "ACTIVE")
	client3DeputyID := suite.seeder.CreateDeputy(ctx, client3ID, "Real", "Deputy", "PRO")
	suite.seeder.CreateAddresses(ctx, client3DeputyID, []string{"93 Fake Avenue"}, "Blompton", "Heartwoodshire", "NO2 RLY", false)
	suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreatePayment(ctx, 15000, yesterday.Date(), "87651234", shared.TransactionTypeMotoCardPayment, today.Date(), 0)

	// client 4 with an unpaid invoice and do not invoice warning
	client4ID := suite.seeder.CreateClient(ctx, "Ian", "McGregor", "12348765", "4132", "ACTIVE")
	client4DeputyID := suite.seeder.CreateDeputy(ctx, client4ID, "Jason", "Statham", "PRO")
	suite.seeder.CreateDeputyImportantInformation(ctx, client4DeputyID, "INVOICE")

	//pfa
	suite.seeder.CreateOrder(ctx, client4ID, "pfa")
	orderId = suite.seeder.GetLastOrderIdForClient(ctx, client4ID)
	suite.seeder.UpdateHowDeputyAppointed(ctx, orderId, "JOINTANDSEVERALLY")
	suite.seeder.CreateDeputyOrder(ctx, client4DeputyID, orderId)

	//hw
	suite.seeder.CreateOrder(ctx, client4ID, "hw")
	orderId = suite.seeder.GetLastOrderIdForClient(ctx, client4ID)
	suite.seeder.UpdateHowDeputyAppointed(ctx, orderId, "SOLE")
	suite.seeder.CreateDeputyOrder(ctx, client4DeputyID, orderId)

	suite.seeder.CreateAddresses(ctx, client4DeputyID, []string{"94 Fake Avenue"}, "Realton", "Nonfictionshire", "OK1 NO2", true)
	suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, nil)
	suite.seeder.CreateWarning(ctx, client4ID, "Do not invoice")

	// client 5 with 1 unpaid invoices
	client5ID := suite.seeder.CreateClient(ctx, "Senor", "Sauron", "88552211", "1111", "ACTIVE")
	client5DeputyID := suite.seeder.CreateDeputy(ctx, client5ID, "Frodo", "Baggins", "PA")
	suite.seeder.CreateAddresses(ctx, client5DeputyID, []string{"123 Fake St", "Baggins View"}, "Hobbiton", "Nonfictionshire", "OK1 NO2", true)
	suite.seeder.CreateInvoice(ctx, client5ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, nil)

	//active hw
	suite.seeder.CreateOrder(ctx, client5ID, "hw")
	orderId1 := suite.seeder.GetLastOrderIdForClient(ctx, client5ID)
	suite.seeder.UpdateHowDeputyAppointed(ctx, orderId1, "JOINTANDSEVERALLY")
	suite.seeder.CreateDeputyOrder(ctx, client5DeputyID, orderId1)

	//active pfa - ensure ordering in the lateral join works
	suite.seeder.CreateOrder(ctx, client5ID, "pfa")
	orderId2 := suite.seeder.GetLastOrderIdForClient(ctx, client5ID)
	suite.seeder.UpdateHowDeputyAppointed(ctx, orderId2, "JOINT")
	suite.seeder.CreateDeputyOrder(ctx, client5DeputyID, orderId2)

	// client 6 with 1 unpaid invoices - ensure can find data when one order is closed
	client6ID := suite.seeder.CreateClient(ctx, "Manny", "Morgoth", "22446688", "2222", "ACTIVE")
	client6DeputyID := suite.seeder.CreateDeputy(ctx, client6ID, "Bilbo", "Baggins", "PRO")
	suite.seeder.CreateInvoice(ctx, client6ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, nil)

	//hw that is active
	suite.seeder.CreateOrder(ctx, client6ID, "hw")
	orderId1 = suite.seeder.GetLastOrderIdForClient(ctx, client6ID)
	suite.seeder.UpdateHowDeputyAppointed(ctx, orderId1, "SEVERAL")
	suite.seeder.CreateDeputyOrder(ctx, client6DeputyID, orderId1)

	//pfa but is closed
	suite.seeder.CreateClosedOrder(ctx, client4ID, today.Date(), "")
	orderId2 = suite.seeder.GetLastOrderIdForClient(ctx, client6ID)
	suite.seeder.UpdateHowDeputyAppointed(ctx, orderId2, "JOINT")
	suite.seeder.CreateDeputyOrder(ctx, client6DeputyID, orderId2)

	// client 7 with 1 unpaid invoices - order not valid as has statusoncaseoverride
	client7ID := suite.seeder.CreateClient(ctx, "Elrond", "TheWise", "22446688", "2222", "ACTIVE")
	client7DeputyID := suite.seeder.CreateDeputy(ctx, client7ID, "Gabriel", "Galadriel", "PRO")
	suite.seeder.CreateInvoice(ctx, client7ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, nil)

	//hw that is active
	suite.seeder.CreateOrder(ctx, client7ID, "pfa")
	orderId = suite.seeder.GetLastOrderIdForClient(ctx, client7ID)
	suite.seeder.UpdateHowDeputyAppointed(ctx, orderId, "SEVERAL")
	suite.seeder.CreateDeputyOrder(ctx, client7DeputyID, orderId)
	suite.seeder.UpdateDeputyOrderStatusOnCaseOverride(ctx, orderId, "SUSPENDED")

	c := Client{suite.seeder.Conn}

	rows, err := c.Run(ctx, NewFeeChase())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 7, len(rows))

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
	assert.Equal(suite.T(), "JOINT", results[0]["Appt_type"], "Appt type - client 1")
	assert.Equal(suite.T(), "SCHEDULE", results[0]["Billing_preference"], "Billing preference - client 1")
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
	assert.Equal(suite.T(), "£190.99", results[0]["Total_debt"], "Total debt - client 1")
	assert.Equal(suite.T(), "B2000001/0", results[0]["Invoice1"], "Invoices - client 1")
	assert.Equal(suite.T(), "£90.99", results[0]["Amount1"], "Invoices - client 1")
	assert.Equal(suite.T(), "AD000001/25", results[0]["Invoice2"], "Invoices - client 1")
	assert.Equal(suite.T(), "£100.00", results[0]["Amount2"], "Invoices - client 1")

	// client 2
	assert.Equal(suite.T(), "87654321", results[1]["Case_no"], "Case name - client 2")
	assert.Equal(suite.T(), "3", results[1]["Client_no"], "Client number - client 2")
	assert.Equal(suite.T(), "", results[1]["Client_title"], "Client title - client 2")
	assert.Equal(suite.T(), "Wallace", results[1]["Client_forename"], "Client forename - client 2")
	assert.Equal(suite.T(), "Gromit", results[1]["Client_surname"], "Client surname - client 2")
	assert.Equal(suite.T(), "No", results[1]["Do_not_chase"], "Do not chase - client 2")
	assert.Equal(suite.T(), "Demanded", results[1]["Payment_method"], "Payment method - client 2")
	assert.Equal(suite.T(), "PA", results[1]["Deputy_type"], "Deputy type - client 2")
	//should not return how deputy appointed on a closed order
	assert.Equal(suite.T(), "", results[1]["Appt_type"], "Appt type - client 2")
	//should not return as not set
	assert.Equal(suite.T(), "", results[1]["Billing_preference"], "Billing preference - client 2")
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
	assert.Equal(suite.T(), "£70.21", results[1]["Total_debt"], "Total debt - client 2")
	assert.Equal(suite.T(), "S2000002/0", results[1]["Invoice1"], "Invoices - client 2")
	assert.Equal(suite.T(), "£70.21", results[1]["Amount1"], "Invoices - client 2")

	// client 4
	assert.Equal(suite.T(), "12348765", results[2]["Case_no"], "Case name - client 3")
	assert.Equal(suite.T(), "7", results[2]["Client_no"], "Client number - client 3")
	assert.Equal(suite.T(), "", results[2]["Client_title"], "Client title - client 3")
	assert.Equal(suite.T(), "Ian", results[2]["Client_forename"], "Client forename - client 3")
	assert.Equal(suite.T(), "McGregor", results[2]["Client_surname"], "Client surname - client 3")
	assert.Equal(suite.T(), "Yes", results[2]["Do_not_chase"], "Do not chase - client 3")
	assert.Equal(suite.T(), "Demanded", results[2]["Payment_method"], "Payment method - client 3")
	assert.Equal(suite.T(), "PRO", results[2]["Deputy_type"], "Deputy type - client 3")
	//should return how deputy appointed on pfa if has pfa and hw order
	assert.Equal(suite.T(), "JOINTANDSEVERALLY", results[2]["Appt_type"], "Appt type - client 3")
	assert.Equal(suite.T(), "INVOICE", results[2]["Billing_preference"], "Billing preference - client 3")
	assert.Equal(suite.T(), "", results[2]["Deputy_no"], "Deputy number - client 3")
	assert.Equal(suite.T(), "Yes", results[2]["Airmail"], "Airmail required - client 3")
	assert.Equal(suite.T(), "", results[2]["Deputy_title"], "Deputy title - client 3")
	assert.Equal(suite.T(), "No", results[2]["Deputy_Welsh"], "Deputy Welsh - client 3")
	assert.Equal(suite.T(), "No", results[2]["Deputy_Large_Print"], "Deputy large print - client 3")
	assert.Equal(suite.T(), "Jason Statham", results[2]["Deputy_name"], "Deputy name - client 3")
	assert.Equal(suite.T(), "", results[2]["Email"], "Deputy email - client 3")
	assert.Equal(suite.T(), "94 Fake Avenue", results[2]["Address1"], "Deputy address 1 - client 3")
	assert.Equal(suite.T(), "", results[2]["Address2"], "Deputy address 2 - client 3")
	assert.Equal(suite.T(), "", results[2]["Address3"], "Deputy address 3 - client 3")
	assert.Equal(suite.T(), "Realton", results[2]["City_Town"], "Deputy town - client 3")
	assert.Equal(suite.T(), "Nonfictionshire", results[2]["County"], "Deputy county - client 3")
	assert.Equal(suite.T(), "OK1 NO2", results[2]["Postcode"], "Deputy postcode - client 3")
	assert.Equal(suite.T(), "£100.00", results[2]["Total_debt"], "Total debt - client 3")
	assert.Equal(suite.T(), "AD000004/25", results[2]["Invoice1"], "Invoices - client 3")
	assert.Equal(suite.T(), "£100.00", results[2]["Amount1"], "Invoices - client 3")

	// client 5
	assert.Equal(suite.T(), "88552211", results[3]["Case_no"], "Case name - client 4")
	assert.Equal(suite.T(), "9", results[3]["Client_no"], "Client number - client 4")
	assert.Equal(suite.T(), "", results[3]["Client_title"], "Client title - client 4")
	assert.Equal(suite.T(), "Senor", results[3]["Client_forename"], "Client forename - client 4")
	assert.Equal(suite.T(), "Sauron", results[3]["Client_surname"], "Client surname - client 4")
	assert.Equal(suite.T(), "No", results[3]["Do_not_chase"], "Do not chase - client 4")
	assert.Equal(suite.T(), "Demanded", results[3]["Payment_method"], "Payment method - client 4")
	assert.Equal(suite.T(), "PA", results[3]["Deputy_type"], "Deputy type - client 4")
	//should return how deputy appointed on pfa if has pfa and hw order
	assert.Equal(suite.T(), "JOINT", results[3]["Appt_type"], "Appt type - client 4")
	assert.Equal(suite.T(), "", results[3]["Billing_preference"], "Billing preference - client 4")
	assert.Equal(suite.T(), "", results[3]["Deputy_no"], "Deputy number - client 4")
	assert.Equal(suite.T(), "Yes", results[3]["Airmail"], "Airmail required - client 4")
	assert.Equal(suite.T(), "", results[3]["Deputy_title"], "Deputy title - client 4")
	assert.Equal(suite.T(), "No", results[3]["Deputy_Welsh"], "Deputy Welsh - client 4")
	assert.Equal(suite.T(), "No", results[3]["Deputy_Large_Print"], "Deputy large print - client 4")
	assert.Equal(suite.T(), "Frodo Baggins", results[3]["Deputy_name"], "Deputy name - client 4")
	assert.Equal(suite.T(), "", results[3]["Email"], "Deputy email - client 4")
	assert.Equal(suite.T(), "123 Fake St", results[3]["Address1"], "Deputy address 1 - client 4")
	assert.Equal(suite.T(), "Baggins View", results[3]["Address2"], "Deputy address 2 - client 4")
	assert.Equal(suite.T(), "", results[3]["Address3"], "Deputy address 3 - client 4")
	assert.Equal(suite.T(), "Hobbiton", results[3]["City_Town"], "Deputy town - client 4")
	assert.Equal(suite.T(), "Nonfictionshire", results[3]["County"], "Deputy county - client 4")
	assert.Equal(suite.T(), "OK1 NO2", results[3]["Postcode"], "Deputy postcode - client 4")
	assert.Equal(suite.T(), "£100.00", results[3]["Total_debt"], "Total debt - client 4")
	assert.Equal(suite.T(), "AD000005/25", results[3]["Invoice1"], "Invoices - client 4")
	assert.Equal(suite.T(), "£100.00", results[3]["Amount1"], "Invoices - client 4")

	// client 6
	assert.Equal(suite.T(), "22446688", results[4]["Case_no"], "Case name - client 5")
	assert.Equal(suite.T(), "11", results[4]["Client_no"], "Client number - client 5")
	assert.Equal(suite.T(), "", results[4]["Client_title"], "Client title - client 5")
	assert.Equal(suite.T(), "Manny", results[4]["Client_forename"], "Client forename - client 5")
	assert.Equal(suite.T(), "Morgoth", results[4]["Client_surname"], "Client surname - client 5")
	assert.Equal(suite.T(), "No", results[4]["Do_not_chase"], "Do not chase - client 5")
	assert.Equal(suite.T(), "Demanded", results[4]["Payment_method"], "Payment method - client 5")
	assert.Equal(suite.T(), "PRO", results[4]["Deputy_type"], "Deputy type - client 5")
	assert.Equal(suite.T(), "JOINT", results[4]["Appt_type"], "Appt type - client 5")
	assert.Equal(suite.T(), "", results[4]["Billing_preference"], "Billing preference - client 5")
	assert.Equal(suite.T(), "", results[4]["Deputy_no"], "Deputy number - client 5")
	assert.Equal(suite.T(), "Bilbo Baggins", results[4]["Deputy_name"], "Deputy name - client 5")
	assert.Equal(suite.T(), "£100.00", results[4]["Total_debt"], "Total debt - client 5")
	assert.Equal(suite.T(), "AD000006/25", results[4]["Invoice1"], "Invoices - client 5")
	assert.Equal(suite.T(), "£100.00", results[4]["Amount1"], "Invoices - client 5")

	// client 7
	assert.Equal(suite.T(), "22446688", results[5]["Case_no"], "Case name - client 6")
	assert.Equal(suite.T(), "13", results[5]["Client_no"], "Client number - client 6")
	assert.Equal(suite.T(), "", results[5]["Client_title"], "Client title - client 6")
	assert.Equal(suite.T(), "Elrond", results[5]["Client_forename"], "Client forename - client 6")
	assert.Equal(suite.T(), "TheWise", results[5]["Client_surname"], "Client surname - client 6")
	assert.Equal(suite.T(), "No", results[5]["Do_not_chase"], "Do not chase - client 6")
	assert.Equal(suite.T(), "Demanded", results[5]["Payment_method"], "Payment method - client 6")
	assert.Equal(suite.T(), "PRO", results[5]["Deputy_type"], "Deputy type - client 6")
	assert.Equal(suite.T(), "", results[5]["Appt_type"], "Appt type - client 6")
	assert.Equal(suite.T(), "", results[5]["Billing_preference"], "Billing preference - client 6")
	assert.Equal(suite.T(), "", results[5]["Deputy_no"], "Deputy number - client 6")
	assert.Equal(suite.T(), "Gabriel Galadriel", results[5]["Deputy_name"], "Deputy name - client 6")
	assert.Equal(suite.T(), "£100.00", results[5]["Total_debt"], "Total debt - client 6")
	assert.Equal(suite.T(), "AD000007/25", results[5]["Invoice1"], "Invoices - client 6")
	assert.Equal(suite.T(), "£100.00", results[5]["Amount1"], "Invoices - client 6")
}
