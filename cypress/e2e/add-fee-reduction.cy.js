describe("Add fee reduction form", () => {
    it("shows correct error message for all potential errors", () => {
        cy.visit("/clients/1/fee-reductions/add");
        cy.get('.govuk-button').click()
        cy.get('.govuk-error-summary').contains("Enter a reason for awarding fee reduction")
        cy.get(".govuk-form-group--error").should('have.length', 5)
    });

    it("shows correct success message", () => {
        cy.visit("/clients/1/fee-reductions?success=fee-reduction[HARDSHIP]");
        cy.get('.moj-banner__message').contains("The hardship has been successfully added")
    });

    it("allows me to enter note information which amends character count", () => {
        cy.visit("/clients/1/fee-reductions/add");
            cy.get(
                ".govuk-character-count > .govuk-form-group > .govuk-label"
            ).should("contain", "Reasons for fee reduction");
            cy.get("#fee-reduction-notes-info + .govuk-character-count__status").should(
                "contain",
                "You have 1,000 characters remaining"
            );
        cy.get("#fee-reduction-notes").type("example note text");
        cy.get("#fee-reduction-notes-info + .govuk-character-count__status").should(
            "contain",
            "You have 983 characters remaining"
        );
    });
});
