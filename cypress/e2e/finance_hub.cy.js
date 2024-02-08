describe("Finance Hub", () => {
    beforeEach(() => {
        cy.setCookie("Other", "other");
        cy.setCookie("XSRF-TOKEN", "abcde");
        cy.visit("/supervision/finance/2");
    });

    describe("Finance Details Header", () => {
        it("should show the finance person name", () => {
            cy.get('#hook-person-name').should("contain", "Finance Person");
        });
        it("should show the finance person name", () => {
            cy.get('#hook-court-ref').should("contain", "12345678");
        });
        it("should show the finance person name", () => {
            cy.get('#hook-total-outstanding-balance').should("contain", "£22.22");
        });
        it("should show the finance person name", () => {
            cy.get('#hook-total-credit-balance').should("contain", "£1.01");
        });
        it("should show the finance person name", () => {
            cy.get('#hook-payment-method').should("contain", "Demanded");
        });
    });
});
