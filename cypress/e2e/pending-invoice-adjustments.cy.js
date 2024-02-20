describe("Pending Invoice Adjustments Tab", () => {
    beforeEach(() => {
        cy.visit("/clients/1/pending-invoice-adjustments");
    });

    describe("Pending Invoice Adjustments", () => {
        it("shows table header", () => {
            cy.contains('[data-cy="invoice"]', "Invoice");
            cy.contains('[data-cy="outstanding"]', "Outstanding invoice balance");
            cy.contains('[data-cy="type"]', "Adjustment type");
            cy.contains('[data-cy="amount"]', "Adjustment amount");
            cy.contains('[data-cy="notes"]', "Notes");
        });
    });
});
