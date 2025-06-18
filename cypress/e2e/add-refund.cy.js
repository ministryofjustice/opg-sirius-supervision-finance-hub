describe("Add Refunds", () => {
    it("navigates from the refunds tab and adds a refund", () => {
        cy.visit("/clients/15/refunds");

        // navigate to form
        cy.contains(".moj-button-menu", "Add refund").click();

        // ensure validation is configured correctly
        cy.contains(".govuk-button", "Save and continue").click();
        cy.get(".govuk-error-summary").contains("Enter the name on the account");
        cy.get(".govuk-error-summary").contains("Enter the account number");
        cy.get(".govuk-error-summary").contains("Enter a reason for the refund");
        cy.get(".govuk-error-summary").contains("Enter the sort code");

        // successfully submit
        cy.contains("label", "Name on bank account").type("Ms Regina Refund");
        cy.contains("label", "Account number").type("12345678");
        cy.contains("label", "Sort code").type("112233"); // will be automatically hyphenated
        cy.contains("label", "Reasons for refund").type("This refund is needed for reasons");
        cy.contains(".govuk-character-count__status", "You have 967 characters remaining");
        cy.contains(".govuk-button", "Save and continue").click();

        cy.url().should("include", "/clients/15/refunds?success=refund-added");

        cy.get(".moj-banner__message").contains("The refund has been successfully added");
    });

    it("should have no accessibility violations",() => {
        cy.visit("/clients/15/refunds/add");
        cy.checkAccessibility();
    });
});
