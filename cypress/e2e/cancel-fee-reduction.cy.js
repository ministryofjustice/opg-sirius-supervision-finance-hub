describe("Cancel fee reduction form", () => {
    it("shows correct error message for all potential errors", () => {
        cy.visit("/clients/1/fee-reductions/1/cancel");
        cy.get('.govuk-button').click()
        cy.get('.govuk-error-summary').contains("Enter a reason for cancelling fee reduction")
        cy.get(".govuk-form-group--error").should('have.length', 1)
    });

    it("shows correct success message", () => {
        cy.visit("/clients/1/fee-reductions?success=fee-reduction[CANCELLED]");
        cy.get('.moj-banner__message').contains("The fee reduction has been successfully cancelled")
    });

    it("allows me to enter note information which amends character count", () => {
        cy.visit("/clients/1/fee-reductions/1/cancel");
        cy.get("#cancel-fee-reduction-notes").type("example note text");
        cy.get("#cancel-fee-reduction-notes-info + .govuk-character-count__status").should(
            "contain",
            "You have 983 characters remaining"
        );
    });
});
