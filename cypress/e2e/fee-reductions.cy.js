describe("Fee Reductions Tab", () => {
    beforeEach(() => {
        cy.visit("/clients/1/fee-reductions");
    });

    describe("Fee Reductions", () => {
        it("displays table and content", () => {
            cy.get("table#fee-reductions > thead > tr")
                .children()
                .first().contains("Type")
                .next().contains("Start date")
                .next().contains("End date")
                .next().contains("Date received")
                .next().contains("Status")
                .next().contains("Notes");

            cy.get("table#fee-reductions > tbody > tr")
                .should("have.length", 4)
                .first()
                .children()
                .first().contains("Remission")
                .next().contains("01/04/2023")
                .next().contains("31/03/2026")
                .next().contains("02/02/2023")
                .next() // status not added yet
                .next().contains("Remission for 2023/2026");
        });
    });
});
