describe("Invoice Tab", () => {
    it("table with correct headers and content", () => {
        cy.visit("/clients/1/invoices");

        cy.get("table#invoices > thead > tr")
            .children()
            .first().contains("Invoice")
            .next().contains("Status")
            .next().contains("Amount")
            .next().contains("Raised")
            .next().contains("Received")
            .next().contains("Outstanding Balance");

        cy.get("table#invoices > tbody > tr")
            .first()
            .children()
            .first().should("contain", "S206666/18")
            .next().should("contain", "Unpaid")
            .next().should("contain", "£320")
            .next().should("contain", "16/03/2018")
            .next().should("contain", "£0")
            .next().should("contain", "£320");
    });

    it("does not show table for no invoices", () => {
        cy.visit("/clients/2/invoices");
        cy.contains('[data-cy="no-invoices"]', "There are no invoices");
    });

    it("displays the ledger allocations for the invoice", () => {
        cy.visit("/clients/1/invoices");
        cy.contains("S206666/18").click();
        cy.contains('[data-cy="ledger-title"]', "Invoice ledger allocations");
        cy.contains('[data-cy="ledger-amount"]', "Amount");
        cy.contains('[data-cy="ledger-received-date"]', "Received date");
        cy.contains('[data-cy="ledger-transaction-type"]', "Transaction type");
        cy.contains('[data-cy="ledger-status"]', "Status");
        cy.get('[data-cy="ledger-amount-data"]').first().contains("12")
        cy.get('[data-cy="ledger-received-date-data"]').first().contains("04/12/2022")
        cy.get('[data-cy="ledger-transaction-type-data"]').first().contains("Manual Credit");
        cy.get('[data-cy="ledger-status-data"]').first().contains("Pending");
    });

    it("displays the supervision levels for the invoice", () => {
        cy.visit("/clients/1/invoices");
        cy.contains("S206666/18").click();
        cy.contains('[data-cy="supervision-title"]', "Supervision level breakdown");
        cy.contains('[data-cy="supervision-level"]', "Supervision level");
        cy.contains('[data-cy="supervision-amount"]', "Amount");
        cy.contains('[data-cy="supervision-from"]', "From");
        cy.contains('[data-cy="supervision-to"]', "To");
        cy.contains('[data-cy="supervision-level-data"]', "General");
        cy.contains('[data-cy="supervision-amount-data"]', "320");
        cy.contains('[data-cy="supervision-from-data"]', "01/04/2022");
        cy.contains('[data-cy="supervision-to-data"]', "31/03/2023");
    });

    it("should have no accessibility violations", () => {
        cy.visit("/clients/1/invoices");
        cy.contains("S206666/18").click();
        cy.checkAccessibility();
    });
});
