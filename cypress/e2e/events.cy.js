describe('Events API Tests', () => {
    const apiUrl = 'http://localhost:8181';
    const notifyUrl = 'http://localhost:3000/v2/notifications/email';
    const reportRequestedTemplateId = "bade69e4-0eb1-4896-a709-bd8f8371a629";
    const user = {
        id: 2,
        roles: ['Finance Reporting']
    };

    it('should generate and upload a report', () => {
        const reportRequest = {
            ReportType: 'AccountsReceivable',
            AccountsReceivableType: 'AgedDebt',
            FromDate: '2023-01-01',
            ToDate: '2023-12-31',
            Email: 'test@example.com'
        };

        cy.generateJWT(user).then((token) => {
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

            cy.request({
                method: 'GET',
                url: notifyUrl
            }).then((response) => {
                expect(response.body).to.have.property('email_address');
                expect(response.body.emailAddress).to.eq(reportRequest.Email);
                expect(response.body).to.have.property('template_id');
                expect(response.body.templateId).to.eq(reportRequestedTemplateId);
            });
        });
    });

    it('should handle report validation errors', () => {
        const reportRequest = {
            ReportType: 'AccountsReceivable',
            AccountsReceivableType: 'AgedDebt',
            FromDate: '2023-01-01',
            ToDate: '2023-12-31',
        };

        cy.generateJWT(user).then((token) => {
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
})