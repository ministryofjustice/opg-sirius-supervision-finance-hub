describe("Refunds E2E", () => {
    const courtRef = "30303000";
    const accountName = "E2E Test User";
    const accountNumber = "12345678";
    const sortCode = "112233";
    const today = new Date();
    const dd = String(today.getDate()).padStart(2, '0');
    const mm = String(today.getMonth() + 1).padStart(2, '0'); //January is 0!
    const yyyy = today.getFullYear();
    const apiUrl = Cypress.env('FINANCE_API_URL') ?? 'http://localhost:8181';
    const user = {
        id: 2,
        roles: ['Finance Reporting']
    };

    // removes all notify emails from json-server to prevent them getting committed
    after(() => {
        cy.request({
            method: 'GET',
            url: notifyUrl
        }).then((response) => {
            response.body.forEach((item) => {
                cy.request({
                    method: 'DELETE',
                    url: `${jsonServerUrl}/clean/${item.id}`,
                    failOnStatusCode: false
                });
            });
        });
    })

    it("issues a refund on excess payment", () => {
        cy.visit("/clients/30/refunds");
        cy.contains('[data-cy="total-outstanding-balance"]', "£0");
        cy.contains('[data-cy="total-credit-balance"]', "£50");
        cy.get(".moj-button-menu").contains("Add refund").click();

        cy.contains("label", "Name on bank account").type(accountName);
        cy.contains("label", "Account number").type(accountNumber);
        cy.contains("label", "Sort code").type(sortCode);
        cy.contains("label", "Reasons for refund").type("This refund is needed for reasons");
        cy.contains(".govuk-button", "Save and continue").click();

        cy.url().should("include", "/clients/30/refunds");
        cy.get(".moj-banner__message").contains("The refund has been successfully added");
    });

    it("issues a approves refund", () => {
        cy.setUser("4"); // set user as refund approver
        cy.visit("/clients/30/refunds");

        cy.get("table#refunds > tbody > tr").first()
            .find(".form-button-menu")
            .contains("Approve").click();

        cy.get("table#refunds > tbody > tr").first().contains("Approved");
    });

    it("downloads approved refunds for processing", () => {
        cy.task('generateJWT', user).then((token) => {
            cy.request({
                method: 'POST',
                url: `${apiUrl}/reports`,
                body: {
                    reportType: "Debt",
                    debtType: "ApprovedRefunds",
                    email: "test@example.com",
                },
                headers: {
                    Authorization: `Bearer ${token}`,
                }
            }).then((response) => {
                expect(response.status).to.eq(201);
            });
        });

        cy.wait(1000); // async process so give it a second to complete

        cy.visit("/clients/30/refunds");
        cy.get("table#refunds > tbody > tr").first().contains("Processing");
    });

    it("fulfills the refund", () => {
        const data =
            "Court reference,Amount,Bank account name,Bank account number,Bank account sort code,Created by,Approved by\n"
            + `${courtRef},50,${accountName},${accountNumber},${sortCode},testuser,testapprover`;

        cy.task('generateJWT', user).then((token) => {
            cy.request({
                method: 'POST',
                url: `${apiUrl}/uploads`,
                body: {
                    uploadType: "FULFILLED_REFUNDS",
                    emailAddress: "test@test.com",
                    fileName: `Fulfilledrefunds_${yyyy}${mm}${dd}.csv`,
                    uploadDate: `${yyyy}-${mm}-${dd}`,
                    data: btoa(data),

                },
                headers: {
                    Authorization: `Bearer ${token}`,
                }
            }).then((response) => {
                expect(response.status).to.eq(200);
            });
        });

        cy.wait(1000); // async process so give it a second to complete

        cy.visit("/clients/30/refunds");
        cy.get("table#refunds > tbody > tr").first().contains("Fulfilled");
        cy.contains('[data-cy="total-outstanding-balance"]', "£0");
        cy.contains('[data-cy="total-credit-balance"]', "£0");
    });

    it("reverses the refund", () => {
        const data = "Court reference,Amount,Bank date (of original refund)\n"
            + `${courtRef},50,${dd}/${mm}/${yyyy}`;

        cy.task('generateJWT', user).then((token) => {
            cy.request({
                method: 'POST',
                url: `${apiUrl}/uploads`,
                body: {
                    uploadType: "REVERSE_FULFILLED_REFUNDS",
                    emailAddress: "test@test.com",
                    fileName: `rejectedrefunds_${yyyy}${mm}${dd}.csv`,
                    uploadDate: `${yyyy}-${mm}-${dd}`,
                    data: btoa(data),
                },
                headers: {
                    Authorization: `Bearer ${token}`,
                }
            }).then((response) => {
                expect(response.status).to.eq(200);
            });
        });

        cy.wait(1000); // async process so give it a second to complete

        cy.visit("/clients/30/refunds");
        cy.contains('[data-cy="total-outstanding-balance"]', "£0");
        cy.contains('[data-cy="total-credit-balance"]', "£50");
    });

    it("displays the refund events in the billing history", () => {
        cy.visit("/clients/30/billing-history");

        cy.get(".moj-timeline__item").should('have.length', 8);

        cy.get(".moj-timeline__item").eq(0).within(() => {
            cy.get(".moj-timeline__title").contains("Refund reversal of £50 received");
            // balance checks
            cy.contains("£50 unallocated");
        });

        cy.get(".moj-timeline__item").eq(1).within(() => {
            cy.contains(".moj-timeline__title", "Refund of £50 created");
            cy.contains(".govuk-list", "£50 refunded");
        });

        cy.get(".moj-timeline__item").eq(2).within(() => {
            cy.contains(".moj-timeline__title", "Refund of £50 fulfilled");
        });

        cy.get(".moj-timeline__item").eq(3).within(() => {
            cy.contains(".moj-timeline__title", "Refund status of approved updated to processing");
            cy.contains(".moj-timeline__date", "Outstanding balance: £0 Credit balance: £50");
        });

        cy.get(".moj-timeline__item").eq(4).within(() => {
            cy.contains(".moj-timeline__title", "Refund status of pending updated to approved");
            cy.contains(".moj-timeline__date", "Outstanding balance: £0 Credit balance: £50");
        });

        cy.get(".moj-timeline__item").eq(5).within(() => {
            cy.contains(".moj-timeline__title", "Pending refund of £50 added");
            cy.contains(".moj-timeline__date", "Outstanding balance: £0 Credit balance: £50");
            cy.contains(".govuk-list", "Notes: This refund is needed for reasons");
        });
    });
});
