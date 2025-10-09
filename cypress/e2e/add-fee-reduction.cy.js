describe("Add fee reduction form", () => {
    it("adds a fee reduction", () => {
        const startYear = new Date().getFullYear();
        // navigate to form

        cy.visit("/clients/2/fee-reductions");
        cy.contains("a", "Award a fee reduction").click();
        // ensure validation is configured correctly

        cy.contains(".govuk-button", "Save and continue").click();
        cy.get(".govuk-error-summary").contains("Enter a reason for awarding fee reduction");
        cy.get(".govuk-form-group--error").should("have.length", 5);
        // successfully submit

        cy.get("#f-FeeType").contains(".govuk-radios__item", "Hardship").click();
        cy.get("#f-StartYear").find("input").type(startYear);
        cy.get("#f-LengthOfAward").contains(".govuk-radios__item", "Three years").click();
        cy.get("#f-DateReceived").find("input").type( `${startYear}-01-01`);
        cy.get("#f-Notes").type("Needs reduction");
        cy.contains(".govuk-character-count__status", "You have 985 characters remaining");
        cy.contains(".govuk-button", "Save and continue").click();

        cy.url().should("include", "/clients/2/fee-reductions?success=fee-reduction[HARDSHIP]");

        cy.get(".moj-banner__message").contains("The hardship has been successfully added");

        // billing history
        cy.visit("/clients/2/billing-history");

        const now = new Date().toLocaleDateString("en-UK");
        cy.get(".moj-timeline__item").first().within((el) => {
            cy.get(".moj-timeline__title").contains("Hardship awarded");
            cy.get(".moj-timeline__byline").contains(`by Ian Admin, ${now}`);
            cy.get(".moj-timeline__date").contains("Outstanding balance: £0 Credit balance: £0");
            cy.get(".govuk-list > li")
                .first().contains(`Start date: 01/04/${startYear}`)
                .next().contains(`End date: 31/03/${startYear + 3}`)
                .next().contains(`Received date: 01/01/${startYear}`)
                .next().contains("Notes: Needs reduction");
        });
    });

    it("should have no accessibility violations",() => {
        cy.visit("/clients/2/fee-reductions/add");
        cy.checkAccessibility();
    });

    it("should not show direct debit button when viewing the add fee reduction form",() => {
        cy.visit("/clients/2/fee-reductions/add");
        cy.get("#direct-debit-button").should('exist');
        cy.get("#direct-debit-button").should('not.be.visible');
    });
});
