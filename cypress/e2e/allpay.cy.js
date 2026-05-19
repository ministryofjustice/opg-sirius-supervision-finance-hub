describe("Allpay end-to-end", () => {
    const apiUrl = Cypress.env('FINANCE_API_URL') ?? 'http://localhost:8181';

    it("sets up Direct Debit mandate and create payment schedule", () => {
        cy.visit("/clients/29/invoices");
        cy.contains(".govuk-button", "Set up Direct Debit").click();
        cy.get("#f-AccountName").contains("Name").type("MR E E2E");
        cy.get("#f-SortCode").contains("Sort code").type("010000");
        cy.get("#f-AccountNumber").contains("number").type("12345678");
        cy.contains(".govuk-button", "Save and continue").click();
        cy.contains('[data-cy="payment-method"]', "Direct Debit");
    });

    it("cancels the Direct Debit mandate", () => {
        cy.visit("/clients/29/invoices");
        cy.contains(".govuk-button", "Cancel Direct Debit").click();
        cy.get("#cancel-direct-debit-form").contains(".govuk-button", "Cancel Direct Debit").click();
    });

    it("displays the events in the billing history", () => {
        cy.visit("/clients/29/billing-history");
        cy.get(".moj-timeline__item").should('have.length', 6);

        cy.get(".moj-timeline__item").eq(0).within(() => {
            cy.get(".moj-timeline__title").contains("Direct Debit Instruction cancelled");
            cy.get(".moj-timeline__byline").contains(`by Ian Admin`);
            cy.contains("Payment method updated to Demanded");
        });

        cy.get(".moj-timeline__item").eq(1).within(() => {
            cy.contains(".moj-timeline__title", "Direct Debit payment of £100 reversed");
            cy.contains(".moj-timeline__byline", `by Colin Case`);
            cy.contains(".govuk-list", "£100 reversed against AD292929/24");
        });

        // the next three events may appear in any order, but in reality the scheduled event would be first, as a
        // schedule cannot be created before a mandate, and payments are scheduled for at least 14 working days in the future
        cy.contains(".moj-timeline__item", "Direct Debit payment of £100 received").within(() => {
            cy.contains(".moj-timeline__byline", `by Colin Case`);
            cy.contains(".govuk-list", "£100 allocated to AD292929/24");
        });

        cy.contains(".moj-timeline__item", "Direct Debit payment scheduled").within(() => {
            cy.contains(".moj-timeline__byline", `by Ian Admin`);
            cy.contains(".govuk-list", "Direct Debit payment for £100 scheduled for");
        });

        cy.get(".moj-timeline__title").contains("Direct Debit Instruction created");
        cy.get(".moj-timeline__byline").contains(`by Ian Admin`);
        cy.contains("Payment method updated to Direct Debit");
    });
});
