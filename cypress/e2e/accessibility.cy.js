import "cypress-axe";

describe("Accessibility test invoices", { tags: "@axe" }, () => {
    before(() => {
        cy.visit('/finance/2');
        cy.url().should('contain', 'finance')
        cy.injectAxe();
    });

    it("Should have no accessibility violations",() => {
        cy.checkA11y();
    });
});
