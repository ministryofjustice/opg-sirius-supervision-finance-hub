const caseManager = "1"
const financeUser = "3"
const financeManager = "4"

describe("Role-based permissions", () => {
    it("checks permissions for use with no finance roles", () => {
        cy.setCookie("x-test-user-id", caseManager);
        cy.visit("/clients/11/invoices");

        cy.contains(".moj-button-menu__item", "Edit payment method").should("not.exist");

        cy.contains(".moj-button-menu__item", "Add manual invoice").should("not.exist");
        cy.contains(".moj-button-menu__item", "Adjust invoice").should("not.exist");

        cy.visit("/clients/11/fee-reductions");
        cy.contains(".govuk-button", "Award fee reduction").should("not.exist");
        cy.contains(".govuk-button", "Cancel").should("not.exist");

        cy.visit("/clients/11/invoice-adjustments");
        cy.contains(".govuk-button", "Approve").should("not.exist");
        cy.contains(".govuk-button", "Reject").should("not.exist");
    });

    it("checks permissions for Finance User role", () => {
        cy.setCookie("x-test-user-id", financeUser);
        cy.visit("/clients/11/invoices");

        cy.contains(".moj-button-menu__item", "Edit payment method").should("exist");

        cy.contains(".moj-button-menu__item", "Add manual invoice").should("not.exist");
        cy.contains(".moj-button-menu__item", "Adjust invoice").should("exist");

        cy.visit("/clients/11/fee-reductions");
        cy.contains(".govuk-button", "Award a fee reduction").should("exist");
        cy.contains(".govuk-button", "Cancel").should("not.exist");

        cy.visit("/clients/11/invoice-adjustments");
        cy.contains(".govuk-button", "Approve").should("not.exist");
        cy.contains(".govuk-button", "Reject").should("not.exist");
    });

    it("checks permissions for Finance Manager role", () => {
        cy.setCookie("x-test-user-id", financeManager);
        cy.visit("/clients/11/invoices");

        cy.contains(".moj-button-menu__item", "Edit payment method").should("not.exist");

        cy.contains(".moj-button-menu__item", "Add manual invoice").should("exist");
        cy.contains(".moj-button-menu__item", "Adjust invoice").should("not.exist");

        cy.visit("/clients/11/fee-reductions");
        cy.contains(".govuk-button", "Award fee reduction").should("not.exist");
        cy.contains(".govuk-button", "Cancel").should("exist");

        cy.visit("/clients/11/invoice-adjustments");
        cy.contains(".govuk-button", "Approve").should("exist");
        cy.contains(".govuk-button", "Reject").should("exist");
    });
});

