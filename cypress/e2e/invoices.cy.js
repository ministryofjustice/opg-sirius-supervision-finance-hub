describe("Invoice Tab", () => {
    beforeEach(() => {
        cy.visit("/clients/1/invoices");
    });

    describe("Invoices", () => {
        it("shows table header", () => {
            cy.contains('[data-cy="invoice"]', "Invoice");
            cy.contains('[data-cy="status"]', "Status");
            cy.contains('[data-cy="amount"]', "Amount");
            cy.contains('[data-cy="raised"]', "Raised");
            cy.contains('[data-cy="received"]', "Received");
            cy.contains('[data-cy="outstanding-balance"]', "Outstanding Balance");
        });
    });

    describe("Invoices ledger allocations", () => {
        it("shows all the correct data", () => {
            cy.get('#invoice-1').click()
            cy.contains('[data-cy="ledger-title"]', "Invoice ledger allocations");
            cy.contains('[data-cy="ledger-amount"]', "Amount");
            cy.contains('[data-cy="ledger-received-date"]', "Received date");
            cy.contains('[data-cy="ledger-transaction-type"]', "Transaction type");
            cy.contains('[data-cy="ledger-status"]', "Status");
        });
    });

    describe("Supervision level breakdown", () => {
        it("shows all the correct data", () => {
            cy.get('#invoice-2').click()
            cy.contains('[data-cy="supervision-title"]', "Supervision level breakdown");
            cy.contains('[data-cy="supervision-level"]', "Supervision level");
            cy.contains('[data-cy="supervision-amount"]', "Amount");
            cy.contains('[data-cy="supervision-from"]', "From");
            cy.contains('[data-cy="supervision-to"]', "To");
        });
    });
});
