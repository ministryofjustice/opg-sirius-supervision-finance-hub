describe("Add manual invoice form", () => {
    it("shows correct error message for all potential errors", () => {
        cy.visit("/clients/1/invoices/add");
        cy.contains(".govuk-button", "Save and continue").click()
        cy.get(".govuk-error-summary").contains("Enter an amount")
        cy.get(".govuk-form-group--error").should("have.length", 6)
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
        cy.contains(".govuk-table__row", "SO000001/99")
    });

    it("should have no accessibility violations",() => {
        cy.visit("/clients/3/invoices/add");
        cy.checkAccessibility();
    });
});
