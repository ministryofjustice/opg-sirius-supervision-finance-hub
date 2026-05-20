describe("Allpay end-to-end", () => {
    const apiUrl = Cypress.env('FINANCE_API_URL') ?? 'http://localhost:8181';
    const user = {
        id: 2,
        roles: ['Finance Reporting']
    };

    it("sets up Direct Debit mandate and create payment schedule", () => {
        cy.visit("/clients/29/invoices");
        cy.contains(".govuk-button", "Set up Direct Debit").click();
        cy.get("#f-AccountName").contains("Name").type("MR E E2E");
        cy.get("#f-SortCode").contains("Sort code").type("010000");
        cy.get("#f-AccountNumber").contains("number").type("12345678");
        cy.contains(".govuk-button", "Save and continue").click();
        cy.contains('[data-cy="payment-method"]', "Direct Debit");
    });

    it("collects scheduled payment", () => {
        const holidayApiUrl = Cypress.env('HOLIDAY_API_URL') ?? 'http://localhost:8080/bank-holidays.json';

        cy.request(holidayApiUrl).then((response) => {
            const bankHolidays = new Set(
                response.body['england-and-wales'].events.map(e => e.date)
            );

            const formattedDate = getCollectionDate(bankHolidays);
            const [day, month, year] = formattedDate.split('/');
            const isoDate = `${year}-${month}-${day}`;

            const csvContent = `9800000000000000000,29292900,100,D,${formattedDate}\n`;
            const base64Data = btoa(csvContent);

            cy.task('generateJWT', user).then((token) => {
                cy.request({
                    method: 'POST',
                    url: `${apiUrl}/uploads`,
                    body: {
                        data: base64Data,
                        emailAddress: "test@example.com",
                        uploadType: "DIRECT_DEBITS_COLLECTIONS",
                        uploadDate: isoDate,
                    },
                    headers: {
                        Authorization: `Bearer ${token}`
                    },
                }).then((response) => {
                    expect(response.status).to.eq(200);
                });
            });
        });

        cy.wait(1000); // async process so give it a second to complete

        cy.visit("/clients/29/invoices");
        cy.get("table#invoices > tbody").contains("AD292929/24")
            .parentsUntil("tr").siblings()
            .first().contains("Closed");
    });

    it("reverses the failed payment", () => {
        const holidayApiUrl = Cypress.env('HOLIDAY_API_URL') ?? 'http://localhost:8080/bank-holidays.json';

        cy.request(holidayApiUrl).then((response) => {
            const bankHolidays = new Set(
                response.body['england-and-wales'].events.map(e => e.date)
            );

            const formattedDate = getCollectionDate(bankHolidays);

            const csvContent = `Court reference,Bank date,Received date,Amount\n29292900,${formattedDate},${formattedDate},100\n`;
            const base64Data = btoa(csvContent);

            cy.task('generateJWT', user).then((token) => {
                cy.request({
                    method: 'POST',
                    url: `${apiUrl}/uploads`,
                    body: {
                        data: base64Data,
                        emailAddress: "test@example.com",
                        uploadType: "FAILED_DIRECT_DEBITS_COLLECTIONS",
                    },
                    headers: {
                        Authorization: `Bearer ${token}`
                    },
                }).then((response) => {
                    expect(response.status).to.eq(200);
                });
            });
        });

        cy.wait(1000); // async process so give it a second to complete

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

    it("cancels the Direct Debit mandate", () => {
        cy.visit("/clients/29/invoices");
        cy.contains(".govuk-button", "Cancel Direct Debit").click();
        cy.get("#cancel-direct-debit-form").contains(".govuk-button", "Cancel Direct Debit").click();
        cy.contains('[data-cy="payment-method"]', "Demanded");
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
            cy.contains(".moj-timeline__byline", `by Ian Admin`);
            cy.contains(".govuk-list", "£100 reversed against AD292929/24");
        });

        // the next three events may appear in any order, but in reality the scheduled event would be first, as a
        // schedule cannot be created before a mandate, and payments are scheduled for at least 14 working days in the future
        cy.contains(".moj-timeline__item", "Direct Debit payment of £100 received").within(() => {
            cy.contains(".moj-timeline__byline", `by Ian Admin`);
            cy.contains(".govuk-list", "£100 allocated to AD292929/24");
        });

        cy.contains(".moj-timeline__item", "Direct Debit payment scheduled").within(() => {
            cy.contains(".moj-timeline__byline", `by Ian Admin`);
            cy.contains(".govuk-list", "Direct Debit payment for £100 scheduled for");
        });

        cy.get(".moj-timeline__title").contains("Direct Debit Instruction created");
        cy.get(".moj-timeline__byline").contains(`by Ian Admin`);
        cy.contains("Payment method updated to Direct Debit");
    });
});

function getCollectionDate(bankHolidays = new Set()) {
    // mirrors production logic: add 14 working days, then find next working day on or after the 24th
    let date = new Date();
    date.setUTCHours(0, 0, 0, 0);

    let workingDaysToAdd = 14;
    while (workingDaysToAdd > 0) {
        date.setDate(date.getDate() + 1);
        if (isWorkingDay(date, bankHolidays)) {
            workingDaysToAdd--;
        }
    }

    // advance to the 24th of the current month, or next month if already past it
    let target = new Date(Date.UTC(date.getFullYear(), date.getMonth(), 24));
    if (date.getDate() > 24) {
        target = new Date(Date.UTC(date.getFullYear(), date.getMonth() + 1, 24));
    }

    // find the next working day on or after the 24th
    while (!isWorkingDay(target, bankHolidays)) {
        target.setDate(target.getDate() + 1);
    }

    return `${String(target.getDate()).padStart(2, '0')}/${String(target.getMonth() + 1).padStart(2, '0')}/${target.getFullYear()}`;
}

function isWorkingDay(date, bankHolidays) {
    const dayOfWeek = date.getDay();
    if (dayOfWeek === 0 || dayOfWeek === 6) return false;
    const dateStr = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
    return !bankHolidays.has(dateStr);
}
