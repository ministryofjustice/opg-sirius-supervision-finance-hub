describe("Adjust invoice form", () => {
    it("adds a manual credit to an invoice", () => {
        cy.visit("/clients/1/invoices");
        cy.contains('.govuk-table__row', 'S206666/18').contains("Adjust invoice").click();

        // ensure validation is configured correctly
        cy.get('.govuk-button').click();
        cy.get('.govuk-error-summary').contains("Select the adjustment type");
        cy.get('.govuk-error-summary').contains("Enter a reason for adjustment");

        cy.get('#error-message__AdjustmentType').contains("Select the adjustment type");
        cy.get('#error-message__AdjustmentNotes').contains("Enter a reason for adjustment");
        cy.get(".govuk-form-group--error").should('have.length', 2);

        // successfully submit credit
        cy.get('#f-AdjustmentType').contains(".govuk-radios__item", "Add credit").click();
        cy.get('#f-AdjustmentNotes').type("manual credit for £100");
        cy.get('#f-Amount').type("10000");
        cy.get('.govuk-button').click();

        // validation for amount
        cy.get('.govuk-error-summary').contains("Amount entered must be equal to or less than £");
        cy.get(".govuk-form-group--error").should('have.length', 1);

        cy.get('#f-Amount').find('input').clear();
        cy.get('#f-Amount').type("100");
        cy.get('.govuk-button').click();

        // navigation and success message
        cy.url().should('include', "clients/1/invoices?success=CREDIT_MEMO");
        cy.get('.moj-banner__message').contains("Manual credit successfully created");
    });
});
