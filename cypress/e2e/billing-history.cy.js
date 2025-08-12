describe("Billing History Tab", () => {
    it("displays the billing history", () => {
        cy.visit("/clients/5/invoices");
        cy.contains("a", "Billing History").click();
        cy.url().should("include", "clients/5/billing-history");

        cy.get(".moj-timeline__item").last().within((el) => {
            cy.get(".moj-timeline__title").contains("AD invoice created for £100");
            cy.get(".moj-timeline__byline").contains("by Tina Test, 06/06/2017");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £100 Credit balance: £0");
            cy.contains(".govuk-link", "AD44444/17").click();
        });

        cy.url().should("include", "clients/5/invoices");
    });

    it("displays refunds events", () => {
        cy.visit("/clients/14/invoices");
        cy.contains("a", "Billing History").click();
        cy.url().should("include", "clients/14/billing-history");

        cy.get(".moj-timeline__item").should('have.length', 17);

        cy.get(".moj-timeline__item").last().within((el) => {
            cy.get(".moj-timeline__title").contains("Pending refund of £123.45 added");
            cy.get(".moj-timeline__byline").contains("by Colin Case, 02/01/2020");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
            cy.get(".moj-timeline__description").contains("Cancelled refund");
        });

        cy.get(".moj-timeline__item").eq(15).within((el) => {
            cy.get(".moj-timeline__title").contains("Refund status of pending updated to approved");
            cy.get(".moj-timeline__byline").contains("by Ian Admin, 03/01/2020");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
        });

        cy.get(".moj-timeline__item").eq(14).within((el) => {
            cy.get(".moj-timeline__title").contains("Refund status of approved updated to processing");
            cy.get(".moj-timeline__byline").contains("by Ian Admin, 04/01/2020");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
        });

        cy.get(".moj-timeline__item").eq(13).within((el) => {
            cy.get(".moj-timeline__title").contains("Refund cancelled");
            cy.get(".moj-timeline__byline").contains("by Ian Admin, 06/01/2020");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
        });

        cy.get(".moj-timeline__item").eq(12).within((el) => {
            cy.get(".moj-timeline__title").contains("Pending refund of £123.44 added");
            cy.get(".moj-timeline__byline").contains("by Colin Case, 02/02/2021");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
            cy.get(".moj-timeline__description").contains("Processing refund");
        });

        cy.get(".moj-timeline__item").eq(11).within((el) => {
            cy.get(".moj-timeline__title").contains("Refund status of pending updated to approved");
            cy.get(".moj-timeline__byline").contains("by Ian Admin, 03/02/2021");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
        });

       cy.get(".moj-timeline__item").eq(10).within((el) => {
            cy.get(".moj-timeline__title").contains("Refund status of approved updated to processing");
            cy.get(".moj-timeline__byline").contains("by Ian Admin, 06/02/2021");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
        });

        cy.get(".moj-timeline__item").eq(9).within((el) => {
            cy.get(".moj-timeline__title").contains("Pending refund of £123.43 added");
            cy.get(".moj-timeline__byline").contains("by Colin Case, 02/03/2022");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
            cy.get(".moj-timeline__description").contains("Rejected refund");
        });

        cy.get(".moj-timeline__item").eq(8).within((el) => {
            cy.get(".moj-timeline__title").contains("Refund status of pending updated to rejected");
            cy.get(".moj-timeline__byline").contains("by Ian Admin, 06/03/2022");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
        });

        cy.get(".moj-timeline__item").eq(7).within((el) => {
            cy.get(".moj-timeline__title").contains("Pending refund of £123.42 added");
            cy.get(".moj-timeline__byline").contains("by Colin Case, 02/04/2023");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
            cy.get(".moj-timeline__description").contains("Approved refund");
        });

        cy.get(".moj-timeline__item").eq(6).within((el) => {
            cy.get(".moj-timeline__title").contains("Refund status of pending updated to approved");
            cy.get(".moj-timeline__byline").contains("by Ian Admin, 06/04/2023");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
       });

      cy.get(".moj-timeline__item").eq(4).within((el) => {
           cy.get(".moj-timeline__title").contains("Pending refund of £123.41 added");
           cy.get(".moj-timeline__byline").contains("by Colin Case, 01/05/2024");
           cy.get(".moj-timeline__date").contains("Outstanding balance: £-100 Credit balance: £30");
           cy.get(".moj-timeline__description").contains("Pending refund");
      });

      cy.get(".moj-timeline__item").eq(3).within((el) => {
            cy.get(".moj-timeline__title").contains("Pending refund of £123.40 added");
            cy.get(".moj-timeline__byline").contains("by Colin Case, 01/06/2025");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £-100 Credit balance: £30");
            cy.get(".moj-timeline__description").contains("Fulfilled refund");
        });

     cy.get(".moj-timeline__item").eq(2).within((el) => {
        cy.get(".moj-timeline__title").contains("Refund status of pending updated to approved");
        cy.get(".moj-timeline__byline").contains("by Ian Admin, 02/06/2025");
        cy.get(".moj-timeline__date").contains("Outstanding balance: £-100 Credit balance: £30");
    });

     cy.get(".moj-timeline__item").eq(1).within((el) => {
        cy.get(".moj-timeline__title").contains("Refund status of approved updated to processing");
        cy.get(".moj-timeline__byline").contains("by Ian Admin, 03/06/2025");
        cy.get(".moj-timeline__date").contains("Outstanding balance: £-100 Credit balance: £30");
    });

        cy.get(".moj-timeline__item").eq(0).within((el) => {
            cy.get(".moj-timeline__title").contains("Refund of £123.40 fulfilled");
            cy.get(".moj-timeline__byline").contains("by Ian Admin, 06/06/2025");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £-100 Credit balance: £30");
       });

    });

    it("no history shows correct message", () => {
        cy.visit("/clients/99/billing-history");
        cy.contains("h2.moj-timeline__title", "No billing history for this client").should("be.visible");
    });


    it("should have no accessibility violations",() => {
        cy.visit("/clients/5/billing-history");
        cy.checkAccessibility();
    });
});