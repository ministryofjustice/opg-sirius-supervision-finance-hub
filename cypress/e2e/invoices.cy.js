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
});
