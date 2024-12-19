describe("Direct debit form", () => {
    it("show correct error message for all present fields with errors", () => {
        cy.visit("/clients/1/direct-debit/add");
        cy.contains(".govuk-button", "Save and continue").click()
        cy.get(".govuk-error-summary").contains("Select who the account holder is")
        cy.get(".govuk-error-summary").contains("Enter the name on the account")
        cy.get(".govuk-error-summary").contains("Enter the account number, must consist of 8 digits")
        cy.get(".govuk-error-summary").contains("Sort code must consist of 6 digits and cannot be all zeros")
        cy.get(".govuk-form-group--error").should("have.length", 4)
    });

    it("should have no accessibility violations",() => {
        cy.visit("/clients/1/direct-debit/add");
        cy.checkAccessibility();
    });
});
