describe("Billing History Tab", () => {
    it("the feed of the billing history show", () => {
        cy.visit("/clients/1/invoices");
        cy.contains('a', 'Billing History').click();
        cy.get('.moj-timeline').first().contains("Write off applied to");
        cy.get('.govuk-link').first().click();
        cy.url().should('include', "clients/1/invoices");
    });

    it("no history shows correct message", () => {
        cy.visit("/clients/2/billing-history");
        cy.contains('h2.moj-timeline__title', 'No billing history for this client').should('be.visible');
    });
});