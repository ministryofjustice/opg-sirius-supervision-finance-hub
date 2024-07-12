describe("Billing History Tab", () => {
    it("the feed of the billing history show", () => {
        cy.visit("/clients/1/billing-history");
        cy.get(".moj-timeline").first().contains("Pending credit memo");
        cy.get(".moj-timeline__date").contains("Outstanding balance: £320 Credit balance: £0");
        cy.get(".govuk-link").first().click();
        cy.url().should("include", "clients/1/invoices");
    });

    it("no history shows correct message", () => {
        cy.visit("/clients/4/billing-history");
        cy.contains("h2.moj-timeline__title", "No billing history for this client").should("be.visible");
    });


    it("should have no accessibility violations",() => {
        cy.visit("/clients/1/billing-history");
        cy.checkAccessibility();
    });
});