import "cypress-axe";

describe("Accessibility test client tasks", { tags: "@axe" }, () => {
    before(() => {
        cy.visit('/');
        cy.url().should('contain', 'finance')
        cy.injectAxe();
    });

    it("Should have no accessibility violations",() => {
        cy.checkA11y();
    });
});
