describe("Customer credit balance", () => {
    it("unapplies excess credit", () => {
        // invoice is partially paid
        cy.visit("/clients/7/invoices");

        // initial balance
        cy.get("#person-info").within(() => {
            cy.contains("Total outstanding balance: £70");
            cy.contains("Total credit balance: £0");
        });

        cy.get('[data-cy="ref"]').should("have.length", 1)
            .contains("AD77777/24").click();
        cy.get("table#ledger-allocations > tbody > tr").should("have.length", 1)
            .contains('[data-cy="ledger-amount-data"]', "£30");

        // check add refund button not showing
        cy.get('[data-cy="refunds"]').click();
        cy.get(".moj-button-menu").should("not.contain", "Add refund");

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
        cy.contains("AD77777/24").click();
        cy.get("table#ledger-allocations > tbody > tr").should("have.length", 3);

        const amounts = ["£30", "£100", "£30"];
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

        // check add refund button now shows
        cy.get('[data-cy="refunds"]').click();
        cy.get(".moj-button-menu").contains("Add refund");

        // check billing history
        cy.visit("/clients/7/billing-history");

        const now = new Date().toLocaleDateString("en-UK");
        cy.get(".moj-timeline__item").first().within((el) => {
            cy.get(".moj-timeline__title").contains("Hardship credit of £100 applied to AD77777/24");
            cy.get(".moj-timeline__byline").contains(`by Ian Admin, ${now}`);
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £30");
            cy.get(".govuk-list > li")
                .first().contains("£100 applied to AD77777/24")
                .next().contains("£30 excess credit unapplied");
        });
    });

    it("reapplies excess credit", () => {
        cy.visit("/clients/7/invoices");

        // confirm balance
        cy.get("#person-info").within(() => {
            cy.contains("Total outstanding balance: £0");
            cy.contains("Total credit balance: £30");
        });

        // add manual invoice
        cy.contains(".govuk-button--secondary", "Add manual invoice").click();
        cy.get('[data-cy="invoice-type"]').select("AD");
        cy.get('[data-cy="raised-date-field-input"]').type("2022-01-01");
        cy.contains(".govuk-button", "Save and continue").click();

        // confirm new balance after credit has been reapplied to new invoice
        cy.get("#person-info").within(() => {
            cy.contains("Total outstanding balance: £70");
            cy.contains("Total credit balance: £0");
        });

        // check billing history
        cy.visit("/clients/7/billing-history");

        const now = new Date().toLocaleDateString("en-UK");
        cy.get(".moj-timeline__item").first().within((el) => {
            cy.get(".moj-timeline__title").contains("£30 reapplied to AD000001/21");
            cy.get(".moj-timeline__byline").contains(`by Ian Admin, ${now}`);
            cy.get(".moj-timeline__date").contains("Outstanding balance: £70 Credit balance: £0");
            cy.get(".govuk-list > li")
                .first().contains("£30 reapplied to AD000001/21");
        });
    });
});
