describe("Customer credit balance", () => {
    it("unapplies excess credit", () => {
        // invoice is partially paid
        cy.visit("/clients/6/invoices");

        // initial balance
        cy.get("#person-info").within(() => {
            cy.contains("Total outstanding balance: £70");
            cy.contains("Total credit balance: £0");
        });

        cy.get('[data-cy="ref"]').should("have.length", 1)
            .contains("AD33333/24").click();
        cy.get("table#ledger-allocations > tbody > tr").should("have.length", 1)
            .contains('[data-cy="ledger-amount-data"]', "£30");

        // apply fee reduction
        cy.get('[data-cy="fee-reductions"]').click();
        cy.contains("a", "Award a fee reduction").click();
        cy.get("#f-FeeType").contains(".govuk-radios__item", "Hardship").click();
        cy.get("#f-StartYear").find("input").type("2024");
        cy.get("#f-LengthOfAward").contains(".govuk-radios__item", "One year").click();
        cy.get("#f-DateReceived").find("input").type("2024-01-01");
        cy.get("#f-Notes").type("Generate CCB excess credit");
        cy.contains(".govuk-button", "Save and continue").click();

        // check ledger entries for fee reduction and unapply
        cy.get('[data-cy="invoices"]').click();
        cy.contains("AD33333/24").click();
        cy.get("table#ledger-allocations > tbody > tr").should("have.length", 3);

        const amounts = ["£-30", "£100", "£30"];
        cy.get('[data-cy="ledger-amount-data"]').each(($el, index) => {
            cy.wrap($el).contains(amounts[index]);
        });

        const statuses = ["Unapplied", "Allocated", "Allocated"];
        cy.get('[data-cy="ledger-status-data"]').each(($el, index) => {
            cy.wrap($el).contains(statuses[index]);
        });

        // confirm new balance
        cy.get("#person-info").within(() => {
            cy.contains("Total outstanding balance: £0");
            cy.contains("Total credit balance: £30");
        });
    });
});
