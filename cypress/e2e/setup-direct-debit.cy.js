describe("Direct debit form", () => {
    it("navigates to and redirects on success with banner", () => {
        cy.visit("/clients/1/invoices");
        cy.contains('[data-cy="payment-method"]', "Demanded");
        cy.contains(".govuk-button", "Set up direct debit").click();
        cy.url().should("include", "/clients/1/direct-debit/setup");
        cy.get("#f-AccountName").contains("Name").type("Mrs Account Holder");
        cy.get("#f-SortCode").contains("Sort code").type("010000");
        cy.get("#f-AccountNumber").contains("number").type("12345678");
        cy.contains(".govuk-button", "Save and continue").click()
        cy.url().should("include", "/clients/1/invoices?success=direct-debit");
        cy.get(".moj-banner__message").contains("The Direct Debit has been set up");
        cy.contains('[data-cy="payment-method"]', "Direct Debit");
    });

    it("shows correct empty error messages for all present fields with errors", () => {
        cy.visit("/clients/2/direct-debit/setup");
        cy.contains(".govuk-button", "Save and continue").click()
        cy.get(".govuk-error-summary").contains("Enter the name on the account")
        cy.get(".govuk-error-summary").contains("Enter the account number")
        cy.get(".govuk-error-summary").contains("Enter the sort code")
        cy.get(".govuk-form-group--error").should("have.length", 3)
    });

    it("shows correct length error messages for all present fields with errors", () => {
        cy.visit("/clients/2/direct-debit/setup");
        cy.get("#f-AccountName").contains("Name").type("Mrs Account Holder");
        cy.get("#f-SortCode").contains("Sort code").type("1");
        cy.get("#f-AccountNumber").contains("number").type("123");
        cy.contains(".govuk-button", "Save and continue").click()
        cy.get(".govuk-error-summary").contains("The account number must consist of 8 digits")
        cy.get(".govuk-error-summary").contains("Sort code must consist of 6 digits in the format 00-00-00")
        cy.get(".govuk-form-group--error").should("have.length", 2)
    });

    it("should have no accessibility violations",() => {
        cy.visit("/clients/2/direct-debit/setup");
        cy.checkAccessibility();
    });

    it("shows error messages from non-form data validation", () => {
        cy.visit("/clients/2/direct-debit/setup");
        cy.get("#f-AccountName").contains("Name").type("Mrs Account Holder");
        cy.get("#f-SortCode").contains("Sort code").type("111111");
        cy.get("#f-AccountNumber").contains("number").type("12345678");
        cy.contains(".govuk-button", "Save and continue").click()
        cy.get(".govuk-error-summary").contains("There is no active fee payer deputy for this client. Please check the client's record before setting up the Direct Debit.")
    });

    it("shows error messages from api validation", () => {
        cy.setPrefer("example=invalid");
        cy.visit("/clients/3/direct-debit/setup");
        cy.get("#f-AccountName").contains("Name").type("Mrs Account Holder");
        cy.get("#f-SortCode").contains("Sort code").type("111111");
        cy.get("#f-AccountNumber").contains("number").type("12345678");

        cy.contains(".govuk-button", "Save and continue").click();

        cy.get(".govuk-error-summary").contains("Direct debit cannot be setup due to an unexpected response from AllPay");
    });
});
