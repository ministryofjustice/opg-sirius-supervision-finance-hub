describe("Direct debit form", () => {
    it("shows correct empty error messages for all present fields with errors", () => {
        cy.visit("/clients/1/direct-debit/add");
        cy.contains(".govuk-button", "Save and continue").click()
        cy.get(".govuk-error-summary").contains("Select the account holder")
        cy.get(".govuk-error-summary").contains("Enter the name on the account")
        cy.get(".govuk-error-summary").contains("Enter the account number")
        cy.get(".govuk-error-summary").contains("Enter the sort code")
        cy.get(".govuk-form-group--error").should("have.length", 4)
    });

    it("shows correct length error messages for all present fields with errors", () => {
        cy.visit("/clients/1/direct-debit/add");
        cy.get("#client").click();
        cy.get("#f-AccountName").contains("Name").type("Mrs Account Holder");
        cy.get("#f-SortCode").contains("Sort code").type("1");
        cy.get("#f-AccountNumber").contains("number").type("123");
        cy.contains(".govuk-button", "Save and continue").click()
        cy.get(".govuk-error-summary").contains("The account number must consist of 8 digits")
        cy.get(".govuk-error-summary").contains("Sort code must consist of 6 digits in the format 00-00-00")
        cy.get(".govuk-form-group--error").should("have.length", 2)
    });

    it("should have no accessibility violations",() => {
        cy.visit("/clients/1/direct-debit/add");
        cy.checkAccessibility();
    });

    it("redirects on success with banner", () => {
        cy.visit("/clients/1/direct-debit/add");
        cy.get("#client").click();
        cy.get("#f-AccountName").contains("Name").type("Mrs Account Holder");
        cy.get("#f-SortCode").contains("Sort code").type("010000");
        cy.get("#f-AccountNumber").contains("number").type("12345678");
        cy.contains(".govuk-button", "Save and continue").click()
        cy.url().should("include", "/clients/1/invoices?success=direct-debit");
        cy.get(".moj-banner__message").contains("The Direct Debit has been set up");
    });
});
