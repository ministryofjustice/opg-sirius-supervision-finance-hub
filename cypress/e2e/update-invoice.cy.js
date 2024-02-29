describe("Adjust invoice form", () => {
        it("shows table headers", () => {
            cy.setCookie("fail-route", "notesError");
            cy.visit("/clients/1/invoices");
            cy.get(':nth-child(1) > :nth-child(7) > .moj-button-menu > .moj-button-menu__wrapper > .govuk-button').click()
            cy.get('.govuk-button').click()
        });
    });
