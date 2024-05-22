describe("Pending Invoice Adjustments Tab", () => {
    it("displays table and content", () => {
        cy.visit("/clients/1/pending-invoice-adjustments");

        cy.get("table#pending-invoice-adjustments > thead > tr")
            .children()
            .first().contains("Invoice")
            .next().contains("Date raised")
            .next().contains("Adjustment type")
            .next().contains("Adjustment amount")
            .next().contains("Notes")
            .next().contains("Status")
            .next().contains("Actions");

        cy.get("table#pending-invoice-adjustments > tbody > tr")
            .first()
            .children()
            .first().contains("S206666/18")
            .next().contains("11/04/2022")
            .next().contains("Credit")
            .next().contains("£12")
            .next().contains("credit adjustment for 12.00")
            .next().contains("Pending");

        cy.get("table#pending-invoice-adjustments > tbody > tr")
            .first()
            .children()
            .last().get(".moj-button-menu")
            .first().contains("Approve")
            .next().contains("Reject");
    });
});
