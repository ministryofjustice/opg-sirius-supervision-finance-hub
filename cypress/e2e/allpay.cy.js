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
        cy.request({
            method: 'POST',
            url: `${apiUrl}/events`,
            body: {
                source: "opg.supervision.infra",
                "detail-type": "scheduled-event",
                detail: {
                    trigger: "direct-debit-collection",
                    override: {
                        date: getCollectionDate(1)
                    }
                }
            },
            headers: {
                Authorization: `Bearer test`
            }
        });

        // send a second time in case it was delayed until the next month
        cy.request({
            method: 'POST',
            url: `${apiUrl}/events`,
            body: {
                source: "opg.supervision.infra",
                "detail-type": "scheduled-event",
                detail: {
                    trigger: "direct-debit-collection",
                    override: {
                        date: getCollectionDate(2)
                    }
                }
            },
            headers: {
                Authorization: `Bearer test`
            }
        });

        cy.wait(1000); // async process so give it a second to complete

        cy.visit("/clients/29/invoices");
        cy.get("table#invoices > tbody").contains("AD292929/24")
            .parentsUntil("tr").siblings()
            .first().contains("Closed");
    });
});


function getCollectionDate(offset) {
    const today = new Date();
    const year = today.getFullYear();
    let month = today.getMonth(); // 0-indexed

    // next month if collection date has already passed
    if (today.getDate() > 24) {
        month++;
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

    return `${collectionDate.getFullYear()}-${String(collectionDate.getMonth() + offset).padStart(2, '0')}-${String(collectionDate.getDate()).padStart(2, '0')}`;
}
