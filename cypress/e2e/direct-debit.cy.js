describe("Payment method form", () => {
    it("should be able to post correctly", () => {
        cy.visit("/clients/1/payment-method/add");
        cy.get('#direct').click()
        cy.contains(".govuk-button", "Save and continue").click()

        cy.get('[data-cy="payment-method"]').should("contain", "Payment method: Direct Debit")
    });

    it("should have no accessibility violations",() => {
        cy.visit("/clients/1/payment-method/add");
        cy.checkAccessibility();
    });
});
