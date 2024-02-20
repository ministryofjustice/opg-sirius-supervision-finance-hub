import "cypress-axe";

describe("Accessibility test invoices", { tags: "@axe" }, () => {
    before(() => {
        cy.visit('/clients/1/invoices');
        cy.url().should('contain', 'invoices')
        cy.injectAxe();
    });

    it("Should have no accessibility violations",() => {
        cy.checkA11y();
    });
});
