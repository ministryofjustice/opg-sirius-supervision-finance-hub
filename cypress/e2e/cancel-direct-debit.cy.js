describe("Direct debit form", () => {
    it("should have no accessibility violations",() => {
        cy.visit("/clients/1/direct-debit/add");
        cy.checkAccessibility();
    });

    it("redirects on success with banner", () => {
        cy.visit("/clients/1/direct-debit/cancel");
        cy.contains(".govuk-button", "Cancel Direct Debit").click()
        cy.url().should("include", "/clients/1/invoices?success=cancel-direct-debit");
        cy.get(".moj-banner__message").contains("The Direct Debit has been cancelled");
    });
});
