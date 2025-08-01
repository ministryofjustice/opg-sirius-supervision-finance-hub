const fs = require('fs');
const yaml = require('yaml');
const axios = require('axios');
const OpenAPIResponseValidator = require('openapi-response-validator').default;

// Load and parse your OpenAPI spec
const specYaml = fs.readFileSync('./allpay.yaml', 'utf8');
const openApiSpec = yaml.parse(specYaml);

async function testFailedPayments() {
    const validator = new OpenAPIResponseValidator({
        responses: openApiSpec.paths['/Customers/{SchemeCode}/Mandates/FailedPayments/{FromDate}/{ToDate}/{Page}'].get.responses,
        components: openApiSpec.components || {},
    });
    try {
        const oneWeekAgo = new Date();
        oneWeekAgo.setDate(oneWeekAgo.getDate() - 7);
        const formatted = oneWeekAgo.toISOString().split('T')[0];

        const today = new Date().toISOString().split('T')[0];
        const response = await axios.get(`https://ddtest.allpay.net/AllpayApi//Customers/OPGB/Mandates/FailedPayments/${formatted}/${today}/1`, {
            headers: {
                Authorization: `Bearer 48F423D6-56FB-EF11-86E8-6045BDFC7A0B`,
            }
        });

        console.log(response.data);

        // Validate the response
        const validationResult = validator.validateResponse(response.status, response.data);

        if (validationResult) {
            console.error('❌ Response validation failed:', validationResult);
        } else {
            console.log('✅ Response is valid according to OpenAPI spec');
        }
    } catch (error) {
        console.error('❌ API call failed:', error.message);
    }
}

async function testModulusCheck() {
    const validator = new OpenAPIResponseValidator({
        responses: openApiSpec.paths['/BankAccounts'].get.responses,
        components: openApiSpec.components || {},
    });
    try {
        const response = await axios.get('https://ddtest.allpay.net/AllpayApi/BankAccounts?sortcode=123456&accountnumber=12345678', {
            headers: {
                Authorization: `Bearer 48F423D6-56FB-EF11-86E8-6045BDFC7A0B`,
            }
        });

        console.log(response.data);

        // Validate the response
        const validationResult = validator.validateResponse(response.status, response.data);

        if (validationResult) {
            console.error('❌ Response validation failed:', validationResult);
        } else {
            console.log('✅ Response is valid according to OpenAPI spec');
        }
    } catch (error) {
        console.error('❌ API call failed:', error.message);
    }
}

testFailedPayments();
testModulusCheck();
