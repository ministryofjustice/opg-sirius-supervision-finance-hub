describe("Fee Reductions Tab", () => {
    it("displays table and content", () => {
        const thisYear = new Date().getFullYear();
        cy.visit("/clients/8/fee-reductions");

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
            .should("have.length", 2)
            .as("rows");

        cy.get("@rows")
            .first().within(() => {
            cy.get("td")
                .first().contains("Hardship")
                .next().contains(`01/04/${thisYear}`)
                .next().contains(`31/03/${thisYear + 1}`)
                .next().contains("01/05/2020")
                .next().contains("Active")
                .next().contains("current reduction")
                .next().contains("Cancel");
        });

        cy.get("@rows")
            .last().within(() => {
            cy.get("td")
                .first().contains("Remission")
                .next().contains("01/04/2019")
                .next().contains("31/03/2020")
                .next().contains("01/05/2019")
                .next().contains("Expired")
                .next().contains("notes")
                .next().not("a");
        });
    });

    it("displays message when there are no fee reductions", () => {
        cy.visit("/clients/99/fee-reductions");

        cy.get("table#fee-reductions > tbody > tr")
            .should("have.length", 1)
            .first()
            .contains("There are no fee reductions");
    });

    it("should have no accessibility violations", () => {
        cy.visit("/clients/8/fee-reductions");
        cy.checkAccessibility();
    });
});
