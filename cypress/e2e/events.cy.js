describe('Events API Tests', () => {
    const apiUrl = 'http://localhost:8080';
    const notifyUrl = 'http://localhost:3000/email';
    const reportRequestedTemplateId = "bade69e4-0eb1-4896-a709-bd8f8371a629";

    it('should generate and upload a report', () => {
        // the request comes from admin, so needs jwt token with user
        const reportRequest = {
            ReportType: 'AccountsReceivable',
            AccountsReceivableType: 'AgedDebt',
            FromDate: '2023-01-01',
            ToDate: '2023-12-31',
            Email: 'test@example.com'
        };

        cy.request({
            method: 'POST', url: `${apiUrl}/reports`, body: reportRequest,
            headers: {
                Authorization: 'Bearer test'
            }
        }).then((response) => {
            expect(response.status).to.eq(200);
        });

        cy.get(notifyUrl)
            .then((response) => {
                expect(response.body).to.have.property('email_address');
                expect(response.body.emailAddress).to.eq(reportRequest.Email);
                expect(response.body).to.have.property('template_id');
                expect(response.body.templateId).to.eq(reportRequestedTemplateId);
            });
    });

    it('should handle report validation errors', () => {
        const reportRequest = {
            ReportType: 'AccountsReceivable',
            AccountsReceivableType: 'AgedDebt',
            FromDate: '2023-01-01',
            ToDate: '2023-12-31',
        };

        cy.request({
            method: 'POST',
            url: `${apiUrl}/generateReport`,
            body: reportRequest,
            failOnStatusCode: false
        }).then((response) => {
            expect(response.status).to.eq(412);
            expect(response.body).to.have.property('validation_errors');
        });
    });
});