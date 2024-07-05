describe("Adjust invoice form", () => {
    it("adds a manual credit to an invoice", () => {
        // navigate to form
        cy.visit("/clients/3/invoices");
        cy.contains('.govuk-table__row', 'S203532/24').contains("Adjust invoice").click();

        // ensure validation is configured correctly
        cy.contains('.govuk-button', "Save and continue").click();
        cy.get('.govuk-error-summary').contains("Select the adjustment type");
        cy.get('.govuk-error-summary').contains("Enter a reason for adjustment");

        cy.get('#error-message__AdjustmentType').contains("Select the adjustment type");
        cy.get('#error-message__AdjustmentNotes').contains("Enter a reason for adjustment");
        cy.get(".govuk-form-group--error").should('have.length', 2);

        // successfully submit credit
        cy.get('#f-AdjustmentType').contains(".govuk-radios__item", "Add credit").click();
        cy.get('#f-AdjustmentNotes').type("manual credit for £100");
        cy.get('#f-Amount').type("10000");
        cy.contains('.govuk-button', "Save and continue").click();

        // validation for amount
        cy.get('.govuk-error-summary').contains("Amount entered must be equal to or less than £");
        cy.get(".govuk-form-group--error").should('have.length', 1);

        cy.get('#f-Amount').find('input').clear();
        cy.get('#f-Amount').type("100");
        cy.contains('.govuk-button', "Save and continue").click();

        // navigation and success message
        cy.url().should('include', "clients/3/invoices?success=invoice-adjustment[CREDIT%20MEMO]");
        cy.get('.moj-banner__message').contains("Manual credit successfully created");
    });

    it("adds debit to an invoice", () => {
        cy.visit("/clients/3/invoices/4/adjustments");

        cy.get('#f-AdjustmentType').contains(".govuk-radios__item", "Add debit").click();
        cy.get('#f-AdjustmentNotes').type("manual debit for £100");
        cy.get('#f-Amount').type("10000");
        cy.contains('.govuk-button', "Save and continue").click();

        // validation for amount
        cy.get('.govuk-error-summary').contains("Amount entered must be equal to or less than £");
        cy.get(".govuk-form-group--error").should('have.length', 1);

        cy.get('#f-Amount').find('input').clear();
        cy.get('#f-Amount').type("100");
        cy.contains('.govuk-button', "Save and continue").click();

        // navigation and success message
        cy.url().should('include', "clients/3/invoices?success=invoice-adjustment[DEBIT%20MEMO]");
        cy.get('.moj-banner__message').contains("Manual debit successfully created");
    });

    it("writes off an invoice", () => {
        cy.visit("/clients/3/invoices/3/adjustments");

        cy.get('#f-AdjustmentType').contains(".govuk-radios__item", "Write off").click();
        cy.get('#f-AdjustmentNotes').type("Writing off");
        cy.get('#f-Amount').should("be.hidden");
        cy.contains('.govuk-button', "Save and continue").click();

        cy.url().should('include', "clients/3/invoices?success=invoice-adjustment[CREDIT%20WRITE%20OFF]");
        cy.get('.moj-banner__message').contains("Write-off successfully created");
    });
});
