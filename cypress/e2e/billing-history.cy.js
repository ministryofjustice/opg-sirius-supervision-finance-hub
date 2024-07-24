describe("Billing History Tab", () => {
    it("the feed of the billing history show", () => {
        cy.visit("/clients/1/invoices");
        cy.contains("a", "Billing History").click();
        cy.url().should("include", "clients/1/billing-history");

        cy.get(".moj-timeline__title").first().contains("Pending credit write off of £12 added to AD03531/19");
        cy.get(".moj-timeline__date").first().contains("Outstanding balance: £520 Credit balance: £0");
        cy.contains(".govuk-link", "AD03531/19").click();
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