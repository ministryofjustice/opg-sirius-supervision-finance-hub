describe("Add fee reduction form", () => {
    it("shows correct error message for all potential errors", () => {
        cy.setCookie("fail-route", "addFeeReductionError");
        cy.visit("/clients/1/fee-reduction/add");
        cy.get('.govuk-button').click()
        cy.get('.govuk-error-summary').contains("A fee reduction type must be selected")
        cy.get('[data-cy="fee-type-error"').contains("A fee reduction type must be selected")
        cy.get(".govuk-form-group--error").should('have.length', 1)
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
