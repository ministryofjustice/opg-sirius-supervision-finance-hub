describe("Cancel fee reduction form", () => {
    it("cancels a fee reduction", () => {
        // navigate to form
        cy.visit("/clients/6/fee-reductions");
        cy.contains("a", "Cancel").click();

        // ensure validation is configured correctly
        cy.contains(".govuk-button", "Save and continue").click();
        cy.get(".govuk-error-summary").contains("Enter a reason for cancelling fee reduction");
        cy.get(".govuk-form-group--error").should("have.length", 1);

        // enter data
        cy.get("#f-CancellationReason").type("Cancelling for reasons");
        cy.contains(".govuk-character-count__status", "You have 978 characters remaining");

        // navigation and success message
        cy.contains(".govuk-button", "Save and continue").click();
        cy.url().should("include", "/clients/6/fee-reductions?success=fee-reduction[CANCELLED]");
        cy.get(".moj-banner__message").contains("The fee reduction has been successfully cancelled");

        // billing history
        cy.visit("/clients/6/billing-history");

        const now = new Date().toLocaleDateString("en-UK");
        cy.get(".moj-timeline__item").first().within((el) => {
            cy.get(".moj-timeline__title").contains("Hardship cancelled");
            cy.get(".moj-timeline__byline").contains(`by Ian Admin, ${now}`);
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
            cy.get(".govuk-list > li")
                .first().contains("Reason: Cancelling for reasons")
        });
    });

    it("should have no accessibility violations",() => {
        cy.visit("/clients/6/fee-reductions/2/cancel");
        cy.checkAccessibility();
    });

    it("should not show Direct Debit button when viewing the cancel fee reduction form",() => {
        cy.visit("/clients/6/fee-reductions/2/cancel");
        cy.get("#direct-debit-button").should('exist');
        cy.get("#direct-debit-button").should('not.be.visible');
    });
});
