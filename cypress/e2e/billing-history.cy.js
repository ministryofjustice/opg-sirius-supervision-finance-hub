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
        cy.visit("/clients/15/invoices");
        cy.contains("a", "Billing History").click();
        cy.url().should("include", "clients/15/billing-history");

        cy.get(".moj-timeline__item").last().within((el) => {
            cy.get(".moj-timeline__title").contains("AD invoice created for £100");
            cy.get(".moj-timeline__byline").contains("by Tina Test, 06/06/2017");
            cy.get(".moj-timeline__date").contains("Outstanding balance: £100 Credit balance: £0");
            cy.contains(".govuk-link", "AD44444/17").click();
        });

        cy.url().should("include", "clients/5/invoices");
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