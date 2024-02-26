describe("Invoice Tab", () => {
        it("shows table headers", () => {
            cy.visit("/clients/1/invoices");
            cy.contains('[data-cy="invoice"]', "Invoice");
            cy.contains('[data-cy="status"]', "Status");
            cy.contains('[data-cy="amount"]', "Amount");
            cy.contains('[data-cy="raised"]', "Raised");
            cy.contains('[data-cy="received"]', "Received");
            cy.contains('[data-cy="outstanding-balance"]', "Outstanding Balance");

            cy.contains('[data-cy="ref"]', "AD101002/22").first();
            cy.contains('[data-cy="invoice-status"]', "Unpaid").first();
            cy.contains('[data-cy="invoice-amount"]', "£123.23").first();
            cy.get(':nth-child(1) > [data-cy="invoice-raised-date"]').contains("10/01/9999");
            cy.contains('[data-cy="invoice-received"]', "£3.45").first();
            cy.contains('[data-cy="invoice-outstanding-balance"]', "£119.78").first();
        });

        it("does not show table for no invoices", () => {
            cy.visit("/clients/2/invoices");
            cy.contains('[data-cy="no-invoices"]', "There are no invoices");
        });
    });

describe("Invoices ledger allocations", () => {
    it("shows all the correct headers", () => {
        cy.visit("/clients/1/invoices");
        cy.get('#invoice-1').click()
        cy.contains('[data-cy="ledger-title"]', "Invoice ledger allocations");
        cy.contains('[data-cy="ledger-amount"]', "Amount");
        cy.contains('[data-cy="ledger-received-date"]', "Received date");
        cy.contains('[data-cy="ledger-transaction-type"]', "Transaction type");
        cy.contains('[data-cy="ledger-status"]', "Status");
        cy.get('[data-cy="ledger-amount-data"]').first().contains("£123")
        cy.get('[data-cy="ledger-received-date-data"]').first().contains("01/05/2222")
        cy.get('[data-cy="ledger-transaction-type-data"]').first().contains("Online card payment");
        cy.get('[data-cy="ledger-status-data"]').first().contains("Applied");
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
        cy.contains('[data-cy="supervision-amount-data"]', "£320").first();
        cy.get('[data-cy="supervision-from-data"]').contains("01/04/2019").first();
        cy.get('[data-cy="supervision-to-data"]').contains("31/03/2020").first();
    });
});