describe("Add manual invoice form", () => {
    it("show correct error message for all present fields with errors", () => {
        cy.visit("/clients/3/invoices/add");
        cy.contains(".govuk-button", "Save and continue").click()
        cy.get(".govuk-error-summary").contains("Please select an invoice type")
        cy.get(".govuk-form-group--error").should("have.length", 1)

        cy.get('[data-cy="invoice-type"]').select("SE");
        cy.get('[data-cy="invoice-type"]').should("have.value", "SE");
        cy.contains(".govuk-button", "Save and continue").click()
        cy.get(".govuk-error-summary").contains("Enter an amount")
        cy.get(".govuk-error-summary").contains("Enter a raised date")
        cy.get(".govuk-error-summary").contains("Enter a start date")
        cy.get(".govuk-error-summary").contains("Enter an end date")
        cy.get(".govuk-error-summary").contains("Please select a valid supervision level")
        cy.get(".govuk-form-group--error").should("have.length", 5)
    });

    it("adds a manual invoice", () => {
        cy.visit("/clients/3/invoices");
        cy.get("a.govuk-button.moj-button-menu__item.govuk-button--secondary").contains("Add manual invoice").click();
        cy.get('[data-cy="invoice-type"]').select("SO");
        cy.get('[data-cy="invoice-type"]').should("have.value", "SO");

        cy.get('[data-cy="amount"]').type("123");
        cy.get('[data-cy="raised-date-field-input"]').type("2024-01-01");
        cy.get('[data-cy="startDate"]').type("9999-01-01");
        cy.get('[data-cy="endDate"]').type("9999-01-01");
        cy.contains("label", "General").click();

        cy.contains(".govuk-button", "Save and continue").click();

        cy.url().should("include", "clients/3/invoices?success=invoice-type[SO]");
        cy.get(".moj-banner__message").contains("The SO invoice has been successfully created");
        cy.contains(".govuk-table__row", "SO000001/98");

        // billing history
        cy.visit("/clients/3/billing-history");

        const now = new Date().toLocaleDateString("en-UK");
        cy.get(".moj-timeline__item").first().within((el) => {
            cy.get(".moj-timeline__title").contains("SO invoice created for £123");
            cy.get(".moj-timeline__byline").contains(`by Ian Admin, ${now}`);
            cy.get(".moj-timeline__date").contains("Outstanding balance: £123 Credit balance: £0");
            cy.contains(".govuk-link", "SO000001/98");
        });
    });

    it("should have no accessibility violations",() => {
        cy.visit("/clients/3/invoices/add");
        cy.checkAccessibility();
    });

    it("should not show Direct Debit button when viewing the add manual invoice form",() => {
        cy.visit("/clients/3/invoices/add");
        cy.get("#direct-debit-button").should('exist');
        cy.get("#direct-debit-button").should('not.be.visible');
    });
});
