describe("Adjust invoice form", () => {
    // TODO: Actual data submission being developed in PFS-29
    // it("shows correct error message for all potential errors", () => {
    //     cy.setCookie("fail-route", "notesError");
    //     cy.visit("/clients/1/invoices");
    //     cy.get(':nth-child(1) > :nth-child(7) > .moj-button-menu > .moj-button-menu__wrapper > .govuk-button').click()
    //     cy.get('.govuk-button').click()
    //     cy.get('.govuk-error-summary').contains("Select the adjustment type")
    //     cy.get('.govuk-error-summary').contains("Reason for manual credit must be 1000 characters or less")
    //     cy.get('.govuk-error-summary').contains("Enter an amount")
    //     cy.get('[data-cy="invoice-error"').contains("Select the adjustment type")
    //     cy.get('[data-cy="notes-error"]').contains("Reason for manual credit must be 1000 characters or less")
    //     cy.get('[data-cy="amount-error"]').contains("Enter an amount")
    //     cy.get(".govuk-form-group--error").should('have.length', 3)
    //
    // });

    it("shows correct success message", () => {
        cy.visit("/clients/1/invoices?success=CREDIT_WRITE_OFF");
        cy.get('.moj-banner__message').contains("The write off is now waiting for approval")
    });
});
