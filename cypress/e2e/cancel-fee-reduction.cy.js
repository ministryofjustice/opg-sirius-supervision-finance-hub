import "cypress-axe";

describe("Cancel fee reduction form", () => {
    it("cancels a fee reduction", () => {
        // navigate to form
        cy.visit("/clients/1/fee-reductions");
        cy.contains('a', "Cancel").click();

        // ensure validation is configured correctly
        cy.contains('.govuk-button', "Save and continue").click();
        cy.get('.govuk-error-summary').contains("Enter a reason for cancelling fee reduction");
        cy.get(".govuk-form-group--error").should('have.length', 1);

        // enter data
        cy.get('#f-Notes').type("Cancelling for reasons");
        cy.get("#cancel-fee-reduction-notes-info + .govuk-character-count__status")
            .contains("You have 978 characters remaining");

        // navigation and success message
        cy.contains('.govuk-button', "Save and continue").click();
        cy.url().should('include', "/clients/1/fee-reductions?success=fee-reduction[CANCELLED]");
        cy.get('.moj-banner__message').contains("The fee reduction has been successfully cancelled")
    });

    it("Should have no accessibility violations",() => {
        // @FIXME
        cy.visit("/clients/2/fee-reductions/3/cancel");
        cy.injectAxe();
        cy.checkA11y();
    });
});
