describe("Finance Hub", () => {
    beforeEach(() => {
        cy.visit("/clients/99/invoices");
    });

    describe("Finance Details Header", () => {
        it("shows the client details", () => {
            cy.contains('[data-cy="person-name"]', "Nelly Nullman");
            cy.contains('[data-cy="court-ref"]', "00000000");
            cy.contains('[data-cy="total-outstanding-balance"]', "£0");
            cy.contains('[data-cy="total-credit-balance"]', "£0");
            cy.contains('[data-cy="payment-method"]', "Demanded");
        });
    });

    describe("Tabs", () => {
        it("navigates between tabs correctly", () => {
            cy.get('[data-cy="fee-reductions"]').click();
            cy.url().should("contain", "fee-reductions");
            cy.contains(".govuk-heading-l", "Fee Reductions");

            cy.get('[data-cy="pending-invoice-adjustments"]').click();
            cy.url().should("contain", "pending-invoice-adjustments");
            cy.contains(".govuk-heading-l", "Pending Invoice Adjustments");

            cy.get('[data-cy="invoices"]').click();
            cy.url().should("contain", "invoices");
            cy.contains(".govuk-heading-l", "Invoices");

            cy.get('[data-cy="billing-history"]').click();
            cy.url().should("contain", "billing-history");
            cy.contains(".govuk-heading-l", "Billing History");
        })
    })
});
