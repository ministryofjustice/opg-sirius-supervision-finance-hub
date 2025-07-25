describe("Refunds tab", () => {
    it("performs actions on refunds", () => {
        cy.visit("/clients/17/refunds");

        cy.get("table#refunds > tbody")
            .contains("Approve me").parent("tr").as("approveRow");

        cy.get("@approveRow")
            .find(".form-button-menu")
            .contains("Approve").click();

        cy.get("@approveRow").contains("Approved");
        cy.get(".moj-banner__message").contains("You have approved the refund");

        cy.get("table#refunds > tbody")
            .contains("Reject me").parent("tr").as("rejectRow");

        cy.get("@rejectRow")
            .find(".form-button-menu")
            .contains("Reject").click();

        cy.get("@rejectRow").contains("Rejected");
        cy.get(".moj-banner__message").contains("You have rejected the refund");
    });
});
