/**
 * Some routes in the API are not accessible via the UI, so we need to test them directly.
 * As this requires the backend, database, and localstack to be running, and they are already
 * running as part of the Cypress suite, we can use the same setup here (despite not being UI-driven tests).
 */
describe('API Tests', () => {
    const apiUrl = Cypress.env('FINANCE_API_URL') ?? 'http://localhost:8181';
    const jsonServerUrl = Cypress.env('JSON_SERVER_URL') ?? 'http://localhost:3000';
    const notifyUrl = `${jsonServerUrl}/v2/notifications/email`;
    const generateReportSuccessTemplateId = "bade69e4-0eb1-4896-a709-bd8f8371a629";
    const processingSuccessTemplateId = "8c85cf6c-695f-493a-a25f-77b4fb5f6a8e";

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

    describe('Report generation', () => {
        const user = {id: 2, roles: ['Finance Reporting']};

        it('should generate and upload a report', () => {
            const reportRequest = {
                ReportType: 'AccountsReceivable',
                AccountsReceivableType: 'AgedDebt',
                FromDate: '2023-01-01',
                ToDate: '2023-12-31',
                Email: 'test@example.com'
            };

            cy.task('generateJWT', user).then((token) => {
                cy.request({
                    method: 'POST',
                    url: `${apiUrl}/reports`,
                    body: reportRequest,
                    headers: {
                        Authorization: `Bearer ${token}`
                    }
                }).then((response) => {
                    expect(response.status).to.eq(201);
                });
            });

            cy.wait(1000); // async process so give it a second to complete

            cy.request({
                method: 'GET',
                url: notifyUrl
            }).then((response) => {
                const notify = response.body.pop();
                expect(notify).to.have.property('email_address');
                expect(notify.email_address).to.eq(reportRequest.Email);
                expect(notify).to.have.property('template_id');
                expect(notify.template_id).to.eq(generateReportSuccessTemplateId);
            });
        });

        it('should handle report validation errors', () => {
            const user = {id: 2, roles: ['Finance Reporting']};
            const reportRequest = {
                ReportType: 'AccountsReceivable',
                AccountsReceivableType: 'AgedDebt',
                FromDate: '2023-01-01',
                ToDate: '2023-12-31',
            };

            cy.task('generateJWT', user).then((token) => {
                cy.request({
                    method: 'POST',
                    url: `${apiUrl}/reports`,
                    body: reportRequest,
                    failOnStatusCode: false,
                    headers: {
                        Authorization: `Bearer ${token}`
                    }
                }).then((response) => {
                    expect(response.status).to.eq(422);
                    expect(response.body).to.have.property('validation_errors');
                });
            });
        });
    });

     describe('Payment processing', () => {
         it('processes payment file from API', () => {
              const user = {
                     id: 2,
                     roles: ['Finance Manager']
                 };

             cy.visit("/clients/13/invoices");
             cy.get("table#invoices > tbody").contains("AD33333/24")
                 .parentsUntil("tr").siblings()
                 .first().contains("Unpaid");

             cy.readFile('fixtures/feemoto_01042025normal.csv', { encoding: 'base64' }).then((base64Data) => {
                 cy.task('generateJWT', user).then((token) => {
                     cy.request({
                         method: 'POST',
                         url: `${apiUrl}/uploads`,
                         body: {
                             data: base64Data,
                             emailAddress: "test@example.com",
                             uploadType: "PAYMENTS_MOTO_CARD",
                             uploadDate: "2025-04-01",
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

             cy.request({
                 method: 'GET',
                 url: notifyUrl
             }).then((response) => {
                 const notify = response.body.pop();
                 expect(notify).to.have.property('email_address');
                 expect(notify.email_address).to.eq("test@example.com");
                 expect(notify).to.have.property('template_id');
                 expect(notify.template_id).to.eq(processingSuccessTemplateId);
             });

             cy.visit("/clients/13/invoices");
             cy.get("table#invoices > tbody").contains("AD33333/24")
                 .parentsUntil("tr").siblings()
                 .first().contains("Closed");
         });
     });

     describe('Direct Debit events', () => {
         it('creates ledgers for payments that have passed their collection date', () => {
             cy.visit("/clients/22/invoices");
             cy.get("table#invoices > tbody").contains("AD222200/24")
                 .parentsUntil("tr").siblings()
                 .first().contains("Unpaid");

             const event = {
                 source: "opg.supervision.infra",
                 "detail-type": "scheduled-event",
                 detail: {
                     trigger: "direct-debit-collection",
                     override: {
                         date: "2025-08-01"
                     }
                 }
             };

             cy.request({
                 method: 'POST',
                 url: `${apiUrl}/events`,
                 body: event,
                 headers: {
                     Authorization: `Bearer test`
                 }
             }).then((response) => {
                 expect(response.status).to.eq(200);
             });

             cy.wait(1000); // async process so give it a second to complete

             cy.visit("/clients/22/invoices");
             cy.get("table#invoices > tbody").contains("AD222200/24")
                 .parentsUntil("tr").siblings()
                 .first().contains("Closed");
         })
     })
});