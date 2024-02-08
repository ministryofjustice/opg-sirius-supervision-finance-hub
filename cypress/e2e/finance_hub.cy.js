describe("Finance Hub", () => {
    beforeEach(() => {
        cy.visit("/supervision/finance/2");
    });

    describe("Finance Details Header", () => {
        it("should shows the person name", () => {
            cy.get('[data-cy="person-name"]').should("contain", "Finance Person");
        });
        it("should shows the court ref", () => {
            cy.get('[data-cy="court-ref"]').should("contain", "12345678");
        });
        it("should shows the total outstanding balance", () => {
            cy.get('[data-cy="total-outstanding-balance"]').should("contain", "£22.22");
        });
        it("should shows the total credit balance", () => {
            cy.get('[data-cy="total-credit-balance"]').should("contain", "£1.01");
        });
        it("should shows the payment method", () => {
            cy.get('[data-cy="payment-method"]').should("contain", "Demanded");
        });
    });
});
