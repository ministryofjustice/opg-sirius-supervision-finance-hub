describe("Adjust invoice form", () => {
    it("adds a manual credit to an invoice", () => {
        cy.visit("/clients/1/invoices");
        cy.get(':nth-child(1) > :nth-child(7) > .moj-button-menu > .moj-button-menu__wrapper > .govuk-button').click();

        // ensure validation is configured correctly
        cy.get('.govuk-button').click();
        cy.get('.govuk-error-summary').contains("Select the adjustment type");
        cy.get('.govuk-error-summary').contains("Reason for manual credit must be 1000 characters or less");
        cy.get('.govuk-error-summary').contains("Enter an amount");

        cy.get('#error-message__adjustmentType').contains("Select the adjustment type");
        cy.get('#error-message__notes').contains("Reason for manual credit must be 1000 characters or less");
        cy.get('#error-message__amount').contains("Enter an amount");
        cy.get(".govuk-form-group--error").should('have.length', 3);

        // successfully submit credit
        cy.get('#f-adjustmentType').contains(".govuk-radios__item", "Add credit").click();
        cy.get('#f-notes').type("manual credit");
        cy.get('#f-amount').type("100");
        cy.get('.govuk-button').click();

        // navigation and success message
        cy.url().contains("/clients/1/invoices?success=CREDIT_MEMO");
        cy.get('.moj-banner__message').contains("Manual credit successfully created");
    });
});
