describe("Invoice Tab", () => {
    beforeEach(() => {
        cy.visit("/supervision/finance/2");
    });

    describe("Invoices", () => {
        it("should shows table header invoice", () => {
            cy.get('[data-cy="invoice"]').should("contain", "Invoice");
        });
        it("should shows table header status", () => {
            cy.get('[data-cy="status"]').should("contain", "Status");
        });
        it("should shows table header amount", () => {
            cy.get('[data-cy="amount"]').should("contain", "Amount");
        });
        it("should shows table header raised", () => {
            cy.get('[data-cy="raised"]').should("contain", "Raised");
        });
        it("should shows table header received", () => {
            cy.get('[data-cy="received"]').should("contain", "Received");
        });
        it("should shows table header outstanding balance", () => {
            cy.get('[data-cy="outstanding-balance"]').should("contain", "Outstanding Balance");
        });
    });
});
