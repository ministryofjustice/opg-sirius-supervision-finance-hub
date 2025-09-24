describe("Adjust invoice form", () => {
    it("adds a manual credit to an invoice", () => {
        // navigate to form
        cy.visit("/clients/4/invoices");
        cy.contains(".govuk-table__row", "AD11111/19").contains("Adjust invoice").click();

        // ensure validation is configured correctly
        cy.contains(".govuk-button", "Save and continue").click();
        cy.get(".govuk-error-summary").contains("Select the adjustment type");
        cy.get(".govuk-error-summary").contains("Enter a reason for adjustment");

        cy.get("#error-message__AdjustmentType").contains("Select the adjustment type");
        cy.get("#error-message__AdjustmentNotes").contains("Enter a reason for adjustment");
        cy.get(".govuk-form-group--error").should("have.length", 2);

        // successfully submit credit
        cy.get("#f-AdjustmentType").contains(".govuk-radios__item", "Add credit").click();
        cy.get("#f-AdjustmentNotes").type("manual credit for £100");
        cy.get("#f-Amount").type("10000");
        cy.contains(".govuk-button", "Save and continue").click();

        // validation for amount
        cy.get(".govuk-error-summary").contains("Amount entered must be equal to or less than £");
        cy.get(".govuk-form-group--error").should("have.length", 1);

        cy.get("#f-Amount").find("input").clear();
        cy.get("#f-Amount").type("100");
        cy.contains(".govuk-button", "Save and continue").click();

        // navigation and success message
        cy.url().should("include", "clients/4/invoices?success=invoice-adjustment[CREDIT%20MEMO]");
        cy.get(".moj-banner__message").contains("Manual credit successfully created");

        // billing history
        cy.visit("/clients/4/billing-history");

        const now = new Date().toLocaleDateString("en-UK");
        cy.get(".moj-timeline__item").first().within((el) => {
            cy.get(".moj-timeline__title").contains("Pending credit memo of £100 added to AD11111/19");
            cy.get(".moj-timeline__byline").contains(`by Ian Admin, ${now}`);
            cy.get(".moj-timeline__date").contains("Outstanding balance: £420 Credit balance: £0");
            cy.contains(".govuk-link", "AD11111/19").click();
        });
    });

    it("writes off an invoice", () => {
        cy.visit("/clients/4/invoices/2/adjustments");

        cy.get("#f-AdjustmentType").contains(".govuk-radios__item", "Write off").click();
        cy.get("#f-AdjustmentNotes").type("Writing off");
        cy.get("#f-Amount").should("be.hidden");
        cy.contains(".govuk-button", "Save and continue").click();

        cy.url().should("include", "clients/4/invoices?success=invoice-adjustment[CREDIT%20WRITE%20OFF]");
        cy.get(".moj-banner__message").contains("Write-off successfully created");
    });

    it("reverses a write off", () => {
        cy.visit("/clients/4/invoices/4/adjustments");

        cy.get("#f-AdjustmentType").contains(".govuk-radios__item", "Write off reversal").click();
        cy.get("#f-AdjustmentNotes").type("Reversing write off");
        cy.get("#f-Amount").should("be.hidden");
        cy.get("#f-manager-override").click();
        cy.get("#f-Amount").type("10");
        cy.contains(".govuk-button", "Save and continue").click();

        cy.url().should("include", "clients/4/invoices?success=invoice-adjustment[WRITE%20OFF%20REVERSAL]");
        cy.get(".moj-banner__message").contains("Write-off reversal successfully created");
    });

    it("does not show manager override checkbox for write off reversals when not a Finance Manager", () => {
        cy.setUser("1");
        cy.visit("/clients/4/invoices/4/adjustments");

        cy.get("#f-AdjustmentType").contains(".govuk-radios__item", "Write off reversal").click();
        cy.get("#f-Amount").should("be.hidden");
        cy.get("#f-manager-override").should("not.exist");
    });

    it("adds debit to an invoice", () => {
        cy.visit("/clients/4/invoices/3/adjustments");

        cy.get("#f-AdjustmentType").contains(".govuk-radios__item", "Add debit").click();
        cy.get("#f-AdjustmentNotes").type("manual debit for £100");
        cy.get("#f-Amount").type("10000");
        cy.contains(".govuk-button", "Save and continue").click();

        // validation for amount
        cy.get(".govuk-error-summary").contains("Amount entered must be equal to or less than £");
        cy.get(".govuk-form-group--error").should("have.length", 1);

        cy.get("#f-Amount").find("input").clear();
        cy.get("#f-Amount").type("100");
        cy.contains(".govuk-button", "Save and continue").click();

        // navigation and success message
        cy.url().should("include", "clients/4/invoices?success=invoice-adjustment[DEBIT%20MEMO]");
        cy.get(".moj-banner__message").contains("Manual debit successfully created");
    });

    it("adds a fee reduction reversal", () => {
        cy.visit("/clients/4/invoices/5/adjustments");

        cy.get("#f-AdjustmentType").contains(".govuk-radios__item", "Fee reduction reversal").click();
        cy.get("#f-AdjustmentNotes").type("Reversing fee reduction");
        cy.get("#f-Amount").type("100");
        cy.contains(".govuk-button", "Save and continue").click();

        // validation
        cy.get(".govuk-error-summary").contains("The fee reduction reversal amount must be £50 or less");

        cy.get("#f-Amount").find("input").clear();
        cy.get("#f-Amount").type("50");
        cy.contains(".govuk-button", "Save and continue").click();

        // navigation and success message
        cy.url().should("include", "clients/4/invoices?success=invoice-adjustment[FEE%20REDUCTION%20REVERSAL]");
        cy.get(".moj-banner__message").contains("Fee reduction reversal successfully created");
    });

    it("should have no accessibility violations",() => {
        cy.visit("/clients/4/invoices/3/adjustments");
        cy.checkAccessibility();
    });
});
