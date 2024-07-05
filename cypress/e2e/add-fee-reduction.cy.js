describe("Add fee reduction form", () => {
    it("adds a fee reduction", () => {
        // navigate to form
        cy.visit("/clients/2/fee-reductions");
        cy.contains('a', "Award a fee reduction").click();

        // ensure validation is configured correctly
        cy.contains('.govuk-button', "Save and continue").click();
        cy.get('.govuk-error-summary').contains("Enter a reason for awarding fee reduction");
        cy.get(".govuk-form-group--error").should('have.length', 5);

        // successfully submit
        cy.get('#f-FeeType').contains(".govuk-radios__item", "Hardship").click();
        cy.get('#f-StartYear').find('input').type("2000c");
        cy.get('#f-LengthOfAward').contains(".govuk-radios__item", "One year").click();
        cy.get('#f-DateReceived').find('input').type("2024-01-01");
        cy.get('#f-Notes').type("Needs reduction");
        cy.get("#fee-reduction-notes-info + .govuk-character-count__status").should(
            "contain",
            "You have 985 characters remaining"
        );
        cy.contains('.govuk-button', "Save and continue").click();
        cy.url().should('include', "/clients/2/fee-reductions?success=fee-reduction[HARDSHIP]");
        cy.get('.moj-banner__message').contains("The hardship has been successfully added");
    });
});
