describe("Billing History Tab", () => {
        it("the feed of the billing history show", () => {
            cy.visit("/clients/1/invoices");
            cy.get('[data-cy="billing-history"]').click()
            cy.get('.moj-timeline').first().contains("Write off applied to");
            cy.get('.govuk-link').first().click();
            cy.url().should('include', "clients/1/invoices");
        });

        it("no history shows correct message", () => {
            cy.visit("/clients/2/billing-history");
            cy.contains('[data-cy="no-billing-history"]', "No billing history for this client");
        });
});