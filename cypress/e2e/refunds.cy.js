describe("Refunds tab", () => {
    it("displays table and content", () => {
        cy.visit("/clients/14/refunds");

        cy.get("table#refunds > thead > tr")
            .children()
            .first().contains("Date raised")
            .next().contains("Date fulfilled")
            .next().contains("Amount")
            .next().contains("Bank details")
            .next().contains("Notes")
            .next().contains("Status")
            .next().contains("Actions");

        cy.get("table#refunds > tbody")
            .contains("Fulfilled").parent("tr").as("row");

        cy.get("@row")
            .children()
            .first().contains("01/06/2025")
            .next().contains("06/06/2025")
            .next().contains("£123.40")
            .next()
            .should($el => {
                expect($el.text().trim()).to.eq('');
            })// no bank details for fulfilled refunds
            .next().contains("Fulfilled refund")
            .next().contains("Fulfilled");
    });

    it("hides bank details and actions from non-finance managers", () => {
        cy.setUser("1");
        cy.visit("/clients/14/refunds");

        cy.get("table#refunds > tbody")
            .contains("Pending").parent("tr").as("pendingRow");

        cy.get("th").should("not.contain", "Bank details");

        cy.get("@pendingRow")
            .find(".form-button-menu")
            .within(() => {
                cy.get("button").should("have.length", 0);
            });
    });

    it("changes visibility based on status", () => {
        cy.visit("/clients/14/refunds");

        cy.get("table#refunds > tbody")
            .contains("Pending").parent("tr").as("pendingRow");

        cy.get("@pendingRow")
            .find(".form-button-menu")
            .within(() => {
                cy.get("button").should("have.length", 2);
                cy.contains("button", "Approve");
                cy.contains("button", "Reject");
            });

        cy.get("table#refunds > tbody")
            .contains("Approved").parent("tr").as("approvedRow");

        cy.get("@approvedRow")
            .children()
            .first()
            .next()
            .next()
            .next()
            .should($el => {
                expect($el.text().trim()).to.eq('');
            }); // bank details in DB but hidden

        cy.get("@approvedRow")
            .find(".form-button-menu")
            .within(() => {
                cy.get("button").should("have.length", 1);
                cy.contains("button", "Cancel");
            });

        cy.get("table#refunds > tbody")
            .contains("Rejected").parent("tr").as("rejectedRow");

        cy.get("@rejectedRow")
            .find(".form-button-menu")
            .within(() => {
                cy.get("button").should("have.length", 0);
            });

        cy.get("table#refunds > tbody")
            .contains("Processing").parent("tr").as("processingRow");

        cy.get("@processingRow")
            .find(".form-button-menu")
            .within(() => {
                cy.get("button").should("have.length", 1);
                cy.contains("button", "Cancel");
            });

        cy.get("table#refunds > tbody")
            .contains("Cancelled").parent("tr").as("cancelledRow");

        cy.get("@cancelledRow")
            .find(".form-button-menu")
            .within(() => {
                cy.get("button").should("have.length", 0);
            });
    });

    it("hides Add Refund button when no credit balance exists", () => {
        cy.visit("/clients/16/refunds");

        cy.get(".moj-button-menu").should("not.contain", "Add refund");
    });

    it("should have no accessibility violations", () => {
        cy.visit("/clients/14/refunds");
        cy.checkAccessibility();
    });
});
