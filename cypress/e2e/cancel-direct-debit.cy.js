describe("Direct debit form", () => {
    it("should have no accessibility violations",() => {
        cy.visit("/clients/19/direct-debit/cancel");
        cy.checkAccessibility();
    });

    it("redirects on success with banner", () => {
        cy.visit("/clients/19/invoices");
        cy.contains('[data-cy="payment-method"]', "Direct Debit");
        cy.contains(".govuk-button", "Cancel direct debit").click();
        cy.url().should("include", "/clients/19/direct-debit/cancel");
        cy.get("#cancel-direct-debit-form").contains(".govuk-button", "Cancel Direct Debit").click();
        cy.url().should("include", "/clients/19/invoices?success=cancel-direct-debit");
        cy.get(".moj-banner__message").contains("The Direct Debit has been cancelled");
        cy.contains('[data-cy="payment-method"]', "Demanded");
    });
});
