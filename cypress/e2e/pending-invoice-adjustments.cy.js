import "cypress-axe";

describe("Pending Invoice Adjustments", () => {
    it("displays table and content", () => {
        cy.visit("/clients/1/pending-invoice-adjustments");

        cy.get("table#pending-invoice-adjustments > thead > tr")
            .children()
            .first().contains("Invoice")
            .next().contains("Date raised")
            .next().contains("Adjustment type")
            .next().contains("Adjustment amount")
            .next().contains("Notes")
            .next().contains("Status")
            .next().contains("Actions");

        cy.get("table#pending-invoice-adjustments > tbody > tr")
            .first()
            .children()
            .first().contains("S206666/18")
            .next().contains("11/04/2022")
            .next().contains("Credit")
            .next().contains("£12")
            .next().contains("credit adjustment for 12.00")
            .next().contains("Pending");

        cy.get("table#pending-invoice-adjustments > tbody > tr")
            .first()
            .children()
            .last().get(".moj-button-menu")
            .first().contains("Approve")
            .next().contains("Reject")
            .click()

        cy.url().should('include', '/pending-invoice-adjustments?success=reject-invoice-adjustment[CREDIT]')

        cy.get('.moj-banner__message').contains("You have rejected the credit")

        cy.get("table#pending-invoice-adjustments > tbody > tr")
            .first()
            .children()
            .last()
            .should("not.contain", ".moj-button-menu")
    });

    it("shows correct success message", () => {
        cy.visit("/clients/1/pending-invoice-adjustments?success=approve-invoice-adjustment[DEBIT]");
        cy.get('.moj-banner__message').contains("You have approved the debit");
        cy.visit("/clients/1/pending-invoice-adjustments?success=reject-invoice-adjustment[WRITE OFF]");
        cy.get('.moj-banner__message').contains("You have rejected the write off");
    });

    it("Should have no accessibility violations",() => {
        cy.visit("/clients/1/pending-invoice-adjustments");
        cy.injectAxe();
        cy.checkA11y();
    });
});
