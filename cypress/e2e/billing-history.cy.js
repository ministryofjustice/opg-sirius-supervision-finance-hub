describe("Billing History Tab", () => {
    it("displays the billing history", () => {
        cy.visit("/clients/5/invoices");
        cy.contains("a", "Billing History").click();
        cy.url().should("include", "clients/5/billing-history");

        cy.get(".moj-timeline__item").last().within(() => {
            cy.get(".moj-timeline__title").contains("AD invoice created for £100");
            cy.get(".moj-timeline__byline").contains("by Tina Test, 06/06/2017");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £100 Credit balance: £0");
            cy.contains(".govuk-link", "AD44444/17").click();
        });

        cy.url().should("include", "clients/5/invoices");
    });

    it("displays payment events", () => {
        cy.visit("/clients/20/invoices");
        cy.contains("a", "Billing History").click();
        cy.url().should("include", "clients/20/billing-history");

        cy.get(".moj-timeline__item").should('have.length', 6);

        cy.get(".moj-timeline__item").last().within(() => {
            cy.get(".moj-timeline__title").contains("AD invoice created for £100");
            cy.get(".moj-timeline__byline").contains("by Ian Admin, 10/04/2024");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £100 Credit balance: £0");
            cy.get(".moj-timeline__description").contains("AD16162/24");
        });

        cy.get(".moj-timeline__item").eq(4).within(() => {
            cy.get(".moj-timeline__title").contains("MOTO card payment of £130 received");
            cy.get(".moj-timeline__byline").contains("by Ian Admin, 11/05/2024");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £30");
            cy.get(".moj-timeline__description").contains("£100 allocated to AD16162/24");
            cy.get(".moj-timeline__description").contains("£30 unallocated");
        });


        cy.get(".moj-timeline__item").eq(3).within(() => {
            cy.get(".moj-timeline__title").contains("AD invoice created for £100");
            cy.get(".moj-timeline__byline").contains("by Ian Admin, 12/05/2024");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £100 Credit balance: £30");
            cy.get(".moj-timeline__description").contains("AD16163/24");
        });

        cy.get(".moj-timeline__item").eq(2).within(() => {
            cy.get(".moj-timeline__title").contains("£30 reapplied to AD16163/24");
            cy.get(".moj-timeline__byline").contains("by Ian Admin, 12/05/2024");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £70 Credit balance: £0");
            cy.get(".moj-timeline__description").contains("£30 reapplied to AD16163/24");
        });

        cy.get(".moj-timeline__item").eq(1).within(() => {
            cy.get(".moj-timeline__title").contains("MOTO card payment of £70 received");
            cy.get(".moj-timeline__byline").contains("by Ian Admin, 13/05/2024");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
            cy.get(".moj-timeline__description").contains("£70 allocated to AD16162/24");
        });

        cy.get(".moj-timeline__item").first().within(() => {
            cy.get(".moj-timeline__title").contains("MOTO card payment of £70 reversed");
            cy.get(".moj-timeline__byline").contains("by Ian Admin, 14/05/2024");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £70 Credit balance: £0");
            cy.get(".moj-timeline__description").contains("£70 reversed against AD16162/24");
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