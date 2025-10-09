describe("Invoice Adjustments", () => {
    it("displays table and content", () => {
        cy.visit("/clients/10/invoice-adjustments");

        cy.get("table#invoice-adjustments > thead > tr")
            .children()
            .first().contains("Invoice")
            .next().contains("Date raised")
            .next().contains("Adjustment type")
            .next().contains("Adjustment amount")
            .next().contains("Notes")
            .next().contains("Status")
            .next().contains("Actions");

        cy.get("table#invoice-adjustments > tbody")
            .contains("AD10101/24").parent("tr").as("row");

        cy.get("@row")
            .children()
            .first().contains("AD10101/24")
            .next().contains("11/04/2022")
            .next().contains("Credit")
            .next().contains("£100")
            .next().contains("credit adjustment for 100.00")
            .next().contains("Pending");

        cy.get("@row")
            .find(".form-button-menu")
            .within(() => {
                cy.contains("button", "Approve");
                cy.contains("button", "Reject");
            });
    });

    it("hides approve button where the adjustment was created by the user", () => {
        cy.setUser("4");
        cy.visit("/clients/10/invoice-adjustments");

        cy.get("table#invoice-adjustments > tbody")
            .contains("AD10101/24").parent("tr")
            .find(".form-button-menu")
            .within(() => {
                cy.contains("button", "Approve").should("not.be.visible");
                cy.contains("button", "Reject");
            });
    });

    it("successfully rejects adjustment", () => {
        cy.visit("/clients/10/invoice-adjustments");

        cy.get("table#invoice-adjustments > tbody")
            .contains("AD10101/24").parent("tr")
            .find(".form-button-menu")
            .within(() => {
                cy.contains("button", "Reject").click();
            });

        cy.url().should("include", "/invoice-adjustments?success=rejected-invoice-adjustment[CREDIT]");

        cy.get(".moj-banner__message").contains("You have rejected the credit");

        cy.get("table#invoice-adjustments > tbody > tr")
            .first()
            .children()
            .last()
            .should("not.contain", ".moj-button-menu");
    });

    it("shows correct success message", () => {
        cy.visit("/clients/10/invoice-adjustments?success=approved-invoice-adjustment[DEBIT]");
        cy.get(".moj-banner__message").contains("You have approved the debit");
        cy.visit("/clients/10/invoice-adjustments?success=rejected-invoice-adjustment[WRITE OFF]");
        cy.get(".moj-banner__message").contains("You have rejected the write off");
    });

    it("should have no accessibility violations",() => {
        cy.visit("/clients/10/invoice-adjustments");
        cy.checkAccessibility();
    });
});
