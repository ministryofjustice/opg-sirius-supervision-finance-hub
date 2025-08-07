describe("Billing History Tab", () => {
//    it("displays the billing history", () => {
//        cy.visit("/clients/5/invoices");
//        cy.contains("a", "Billing History").click();
//        cy.url().should("include", "clients/5/billing-history");
//
//        cy.get(".moj-timeline__item").last().within((el) => {
//            cy.get(".moj-timeline__title").contains("AD invoice created for £100");
//            cy.get(".moj-timeline__byline").contains("by Tina Test, 06/06/2017");
//            cy.get(".moj-timeline__date").contains("Outstanding balance: £100 Credit balance: £0");
//            cy.contains(".govuk-link", "AD44444/17").click();
//        });
//
//        cy.url().should("include", "clients/5/invoices");
//    });

    it("displays refunds events", () => {
        cy.visit("/clients/15/invoices");
        cy.contains("a", "Billing History").click();
        cy.url().should("include", "clients/15/billing-history");

//        pending refund
        cy.get(".moj-timeline__item").should('have.length', 13);

        cy.get(".moj-timeline__item").last().within((el) => {
            cy.get(".moj-timeline__title").contains("Pending refund of £123.45 added");
            cy.get(".moj-timeline__byline").contains("by Colin Case, 01/06/2020");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
            cy.get(".moj-timeline__description").contains("Cancelled refund");
        });

        cy.get(".moj-timeline__item").eq(11).within((el) => {
            cy.get(".moj-timeline__title").contains("Refund cancelled");
            cy.get(".moj-timeline__byline").contains("by Colin Case, 09/06/2020");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
        });

        cy.get(".moj-timeline__item").eq(10).within((el) => {
            cy.get(".moj-timeline__title").contains("Pending refund of £123.44 added");
            cy.get(".moj-timeline__byline").contains("by Colin Case, 01/05/2021");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
            cy.get(".moj-timeline__description").contains("Processing refund");
        });

        cy.get(".moj-timeline__item").eq(9).within((el) => {
            cy.get(".moj-timeline__title").contains("Refund status of approved updated to processing");
            cy.get(".moj-timeline__byline").contains("by Colin Case, 04/05/2021");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
        });

        cy.get(".moj-timeline__item").eq(8).within((el) => {
            cy.get(".moj-timeline__title").contains("Pending refund of £123.43 added");
            cy.get(".moj-timeline__byline").contains("by Colin Case, 01/04/2022");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
            cy.get(".moj-timeline__description").contains("Rejected refund");
        });

        cy.get(".moj-timeline__item").eq(7).within((el) => {
            cy.get(".moj-timeline__title").contains("Refund status of pending updated to rejected");
            cy.get(".moj-timeline__byline").contains("by Colin Case, 06/04/2022");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
        });

        cy.get(".moj-timeline__item").eq(6).within((el) => {
            cy.get(".moj-timeline__title").contains("Pending refund of £123.42 added");
            cy.get(".moj-timeline__byline").contains("by Colin Case, 01/03/2023");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
            cy.get(".moj-timeline__description").contains("Approved refund");
        });

        cy.get(".moj-timeline__item").eq(5).within((el) => {
            cy.get(".moj-timeline__title").contains("Refund status of pending updated to approved");
            cy.get(".moj-timeline__byline").contains("by Colin Case, 06/03/2023");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
       });

      cy.get(".moj-timeline__item").eq(4).within((el) => {
           cy.get(".moj-timeline__title").contains("Pending refund of £123.41 added");
           cy.get(".moj-timeline__byline").contains("by Colin Case, 01/02/2024");
           cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
           cy.get(".moj-timeline__description").contains("Pending refund");
      });

      cy.get(".moj-timeline__item").eq(1).within((el) => {
            cy.get(".moj-timeline__title").contains("Pending refund of £123.40 added");
            cy.get(".moj-timeline__byline").contains("by Colin Case, 01/01/2025");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £30");
            cy.get(".moj-timeline__description").contains("Fulfilled refund");
        });

        cy.get(".moj-timeline__item").eq(0).within((el) => {
            cy.get(".moj-timeline__title").contains("Refund of £123.40 fulfilled");
            cy.get(".moj-timeline__byline").contains("by Colin Case, 08/01/2025");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £123.40 Credit balance: £30");
       });

    });
//
//    it("no history shows correct message", () => {
//        cy.visit("/clients/99/billing-history");
//        cy.contains("h2.moj-timeline__title", "No billing history for this client").should("be.visible");
//    });
//
//
//    it("should have no accessibility violations",() => {
//        cy.visit("/clients/5/billing-history");
//        cy.checkAccessibility();
//    });
});