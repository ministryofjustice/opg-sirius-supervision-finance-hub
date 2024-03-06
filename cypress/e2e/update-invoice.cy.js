describe("Adjust invoice form", () => {
    it("shows correct error message", () => {
        cy.setCookie("fail-route", "notesError");
        cy.visit("/clients/1/invoices");
        cy.get(':nth-child(1) > :nth-child(7) > .moj-button-menu > .moj-button-menu__wrapper > .govuk-button').click()
        cy.get('.govuk-button').click()
        cy.get('.govuk-error-summary').contains("The note must be 1000 characters or fewer")
    });

    it("shows correct success message", () => {
        cy.setCookie("success-route", "/invoices?clientId=1?");
        cy.visit("/clients/1/invoices");
        cy.get(':nth-child(1) > :nth-child(7) > .moj-button-menu > .moj-button-menu__wrapper > .govuk-button').click()
        cy.get('#writeOff').check({force:true});
        cy.get('.govuk-button').click()
        cy.get('.moj-banner__message').contains("The write off is now waiting for approval")
    });
});
