describe("Footer", () => {
    beforeEach(() => {
        cy.visit("/clients/99/invoices");
    });

    it("should show the accessibility link", () => {
        cy.get('[data-cy="accessibilityStatement"]').should("contain", "Accessibility statement");
    });
});
