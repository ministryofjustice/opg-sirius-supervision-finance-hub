describe("Add fee reduction form", () => {
    it("shows correct error message for all potential errors", () => {
        cy.setCookie("fail-route", "addFeeReductionError");
        cy.visit("/clients/1/fee-reductions/add");
        cy.get('.govuk-button').click()
        cy.get('.govuk-error-summary').contains("Enter a reason for awarding fee reduction")
        cy.get(".govuk-form-group--error").should('have.length', 4)
    });

    it("shows correct success message", () => {
        cy.visit("/clients/1/fee-reductions?success=hardship");
        cy.get('.moj-banner__message').contains("The hardship has been successfully add")
    });
});
