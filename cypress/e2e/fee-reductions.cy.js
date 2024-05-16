describe("Fee Reductions Tab", () => {
    it("displays table and content", () => {
        cy.visit("/clients/1/fee-reductions");

        cy.get("table#fee-reductions > thead > tr")
            .children()
            .first().contains("Type")
            .next().contains("Start date")
            .next().contains("End date")
            .next().contains("Date received")
            .next().contains("Status")
            .next().contains("Reasons for fee reduction")
            .next().contains("Actions");

        cy.get("table#fee-reductions > tbody > tr")
            .should("have.length", 1)
            .first()
            .children()
            .first().contains("Remission")
            .next().contains("01/04/2019")
            .next().contains("31/03/2020")
            .next().contains("01/05/2019")
            .next().contains("Expired")
            .next().contains("notes");
    });

    it("displays message when there are no fee reductions", () => {
        cy.visit("/clients/2/fee-reductions");

        cy.get("table#fee-reductions > tbody > tr")
            .should("have.length", 1)
            .first()
            .contains("There are no fee reductions");
    });
});
