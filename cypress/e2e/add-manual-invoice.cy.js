describe("Add manual invoice form", () => {
    it("shows correct error message for all potential errors", () => {
        cy.setCookie("fail-route", "addManualInvoiceError");
        cy.visit("/clients/1/manual-invoice");
        cy.get('.govuk-button').click()
        cy.get('.govuk-error-summary').contains("Enter an amount")
        cy.get(".govuk-form-group--error").should('have.length', 4)
    });

    it("adds a manual invoice", () => {
        cy.visit("/clients/3/invoices");
        cy.get('a.govuk-button.moj-button-menu__item.govuk-button--secondary').contains("Add manual invoice").click();
        cy.get('#invoice-type').select('SO');
        cy.get('#invoice-type').should('have.value', 'SO');

        cy.get('#amount').type("123");
        cy.get('#raised-date-field-input').type("2024-01-01");
        cy.get('#startDate').type("9999-01-01");
        cy.get('#endDate').type("9999-01-01");
        cy.contains('label', 'General').click();

        cy.get('.govuk-button').click();

        cy.url().should('include', "clients/3/invoices?success=invoice-type[SO]");
        cy.get('.moj-banner__message').contains("The SO type has been successfully created");
    });
});
