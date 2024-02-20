describe("Fee Reductions Tab", () => {
    beforeEach(() => {
        cy.visit("/clients/1/fee-reductions");
    });

    describe("Fee Reductions", () => {
        it("shows table header", () => {
            cy.contains('[data-cy="type"]', "Type");
            cy.contains('[data-cy="start-date"]', "Start date");
            cy.contains('[data-cy="end-date"]', "End date");
            cy.contains('[data-cy="date-received"]', "Date received");
            cy.contains('[data-cy="status"]', "Status");
            cy.contains('[data-cy="notes"]', "Notes");
        });
    });
});
