describe("Add fee reduction form", () => {
    it("shows correct error message for all potential errors", () => {
        cy.setCookie("fail-route", "addFeeReductionError");
        cy.visit("/clients/1/fee-reductions/add");
        cy.get('.govuk-button').click()
        cy.get('.govuk-error-summary').contains("Date received must be in the past")
        cy.get(".govuk-form-group--error").should('have.length', 5)
    });

    // it("shows correct success message", () => {
    //     cy.setCookie("success-route", "/invoices?clientId=1?");
    //     cy.visit("/clients/1/invoices");
    //     cy.get(':nth-child(1) > :nth-child(7) > .moj-button-menu > .moj-button-menu__wrapper > .govuk-button').click()
    //     cy.get('#writeOff').check({force:true});
    //     cy.get('.govuk-button').click()
    //     cy.get('.moj-banner__message').contains("The write off is now waiting for approval")
    //     cy.url().should('include', '/clients/1/invoices?success=writeOff')
    // });
});
