describe("Allpay end-to-end", () => {
    const apiUrl = Cypress.env('FINANCE_API_URL') ?? 'http://localhost:8181';

    it("sets up direct debit mandate and create payment schedule", () => {
        cy.visit("/clients/29/invoices");
        cy.contains(".govuk-button", "Set up direct debit").click();
        cy.get("#f-AccountName").contains("Name").type("MR E E2E");
        cy.get("#f-SortCode").contains("Sort code").type("010000");
        cy.get("#f-AccountNumber").contains("number").type("12345678");
        cy.contains(".govuk-button", "Save and continue").click();
        cy.contains('[data-cy="payment-method"]', "Direct Debit");
    });

    it("collects scheduled payment", () => {
        const sendEvent = (date) => cy.request({
            method: 'POST',
            url: `${apiUrl}/events`,
            body: {
                source: "opg.supervision.infra",
                "detail-type": "scheduled-event",
                detail: {
                    trigger: "direct-debit-collection",
                    override: {
                        date: date
                    }
                }
            },
            headers: {
                Authorization: `Bearer test`
            }
        });

        sendEvent(getCollectionDate(0));
        // send a second time in case it was delayed until the next month
        sendEvent(getCollectionDate(1));

        cy.wait(1000); // async process so give it a second to complete

        cy.visit("/clients/29/invoices");
        cy.get("table#invoices > tbody").contains("AD292929/24")
            .parentsUntil("tr").siblings()
            .first().contains("Closed");
    });

    it("reverses the failed payment", () => {
        cy.request({
            method: 'POST',
            url: `${apiUrl}/events`,
            body: {
                source: "opg.supervision.infra",
                "detail-type": "scheduled-event",
                detail: {
                    trigger: "failed-direct-debit-collections",
                    override: {
                        date: "2000-01-01"
                    }
                }
            },
            headers: {
                Authorization: `Bearer test`
            }
        });

        cy.visit("/clients/29/invoices");

        cy.get("table#invoices > tbody").contains("AD292929/24").as("invoice");
        cy.get("@invoice").click();

        cy.get("@invoice").parentsUntil("tr").siblings()
            .first().contains("Unpaid");

        cy.get("table#ledger-allocations > tbody")
            .within(el => {
                cy.get("tr").should('have.length', 2);
            });
    });

    it("cancels the direct debit mandate", () => {
        cy.visit("/clients/29/invoices");
        cy.contains(".govuk-button", "Cancel direct debit").click();
        cy.get("#cancel-direct-debit-form").contains(".govuk-button", "Cancel Direct Debit").click();
    });

    it("displays the events in the billing history", () => {
        cy.visit("/clients/29/billing-history");
        cy.get(".moj-timeline__item").should('have.length', 6);

        cy.get(".moj-timeline__item").eq(0).within(() => {
            cy.get(".moj-timeline__title").contains("Direct Debit Instruction cancelled");
            cy.get(".moj-timeline__byline").contains(`by Ian Admin`);
            cy.contains("Payment method updated to Demanded");
        });

        cy.get(".moj-timeline__item").eq(1).within(() => {
            cy.contains(".moj-timeline__title", "Direct Debit payment of £100 reversed");
            cy.contains(".moj-timeline__byline", `by Colin Case`);
            cy.contains(".govuk-list", "£100 reversed against AD292929/24");
        });

        // the next three events may appear in any order, but in reality the scheduled event would be first, as a
        // schedule cannot be created before a mandate, and payments are scheduled for at least 14 working days in the future
        cy.contains(".moj-timeline__item", "Direct Debit payment of £100 received").within(() => {
            cy.contains(".moj-timeline__byline", `by Colin Case`);
            cy.contains(".govuk-list", "£100 allocated to AD292929/24");
        });

        cy.contains(".moj-timeline__item", "Direct debit payment scheduled").within(() => {
            cy.contains(".moj-timeline__byline", `by Ian Admin`);
            cy.contains(".govuk-list", "Direct debit payment for £100 scheduled for");
        });

        cy.get(".moj-timeline__title").contains("Direct Debit Instruction created");
        cy.get(".moj-timeline__byline").contains(`by Ian Admin`);
        cy.contains("Payment method updated to Direct debit");
    });
});


function getCollectionDate(offset) {
    const today = new Date();
    let year = today.getFullYear();
    let month = today.getMonth() + offset; // 0-indexed

    // next month if collection date has already passed
    if (today.getDate() > 24) {
        month++;
    }

    if (month > 11) {
        month = month % 12;
        year++;
    }

    let collectionDate = new Date(year, month, 24);

    // Get the day of the week (0 = Sunday, 6 = Saturday)
    const dayOfWeek = collectionDate.getDay();

    // If it's Saturday (6), move to Monday (25th)
    if (dayOfWeek === 6) {
        collectionDate.setDate(26);
    }
    // If it's Sunday (0), move to Monday (25th)
    else if (dayOfWeek === 0) {
        collectionDate.setDate(25);
    }

    return `${collectionDate.getFullYear()}-${String(collectionDate.getMonth() + 1).padStart(2, '0')}-${String(collectionDate.getDate()).padStart(2, '0')}`;
}
