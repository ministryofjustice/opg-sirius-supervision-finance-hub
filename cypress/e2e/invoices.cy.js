describe("Invoice Tab", () => {
        it("table with correct headers and content", () => {
            cy.visit("/clients/1/invoices");
            cy.contains('[data-cy="invoice"]', "Invoice");
            cy.contains('[data-cy="status"]', "Status");
            cy.contains('[data-cy="amount"]', "Amount");
            cy.contains('[data-cy="raised"]', "Raised");
            cy.contains('[data-cy="received"]', "Received");
            cy.contains('[data-cy="outstanding-balance"]', "Outstanding Balance");
            cy.contains('[data-cy="ref"]', "AD03531/19");
            cy.contains('[data-cy="invoice-amount"]', "£100");
            cy.contains('[data-cy="invoice-raised-date"]',"16/03/2020");
            cy.contains('[data-cy="invoice-received"]', "£88");
            cy.contains('[data-cy="invoice-outstanding-balance"]', "£12");
        });

        it("does not show table for no invoices", () => {
            cy.visit("/clients/2/invoices");
            cy.contains('[data-cy="no-invoices"]', "There are no invoices");
        });
    });

describe("Invoices ledger allocations", () => {
    it("table with correct headers and content", () => {
        cy.visit("/clients/1/invoices");
        cy.get('#invoice-2').click()
        cy.contains('[data-cy="ledger-title"]', "Invoice ledger allocations");
        cy.contains('[data-cy="ledger-amount"]', "Amount");
        cy.contains('[data-cy="ledger-received-date"]', "Received date");
        cy.contains('[data-cy="ledger-transaction-type"]', "Transaction type");
        cy.contains('[data-cy="ledger-status"]', "Status");
        cy.get('[data-cy="ledger-amount-data"]').first().contains("£88")
        cy.get('[data-cy="ledger-received-date-data"]').first().contains("04/12/2022")
        cy.get('[data-cy="ledger-transaction-type-data"]').first().contains("Card Payment");
        cy.get('[data-cy="ledger-status-data"]').first().contains("Approved");
    });
});

describe("Supervision level breakdown", () => {
    it("shows all the correct headers", () => {
        cy.visit("/clients/1/invoices");
        cy.get('#invoice-2').click()
        cy.contains('[data-cy="supervision-title"]', "Supervision level breakdown");
        cy.contains('[data-cy="supervision-level"]', "Supervision level");
        cy.contains('[data-cy="supervision-amount"]', "Amount");
        cy.contains('[data-cy="supervision-from"]', "From");
        cy.contains('[data-cy="supervision-to"]', "To");
        cy.contains('[data-cy="supervision-level-data"]', "General").first();
        cy.contains('[data-cy="supervision-amount-data"]', "£100").first();
        cy.contains('[data-cy="supervision-from-data"]', "01/04/2022").first();
        cy.contains('[data-cy="supervision-to-data"]', "31/03/2023").first();
    });
});
