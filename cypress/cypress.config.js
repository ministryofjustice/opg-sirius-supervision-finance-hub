const {defineConfig} = require("cypress");
const { SignJWT } = require('jose');

module.exports = defineConfig({
    fixturesFolder: false,
    e2e: {
        setupNodeEvents(on, config) {
            on("task", {
                log(message) {
                    console.log(message);

                    return null
                },
                table(message) {
                    console.table(message);

                    return null
                },
                async generateJWT(user) {
                    const secret = new TextEncoder().encode(
                        'mysupersecrettestkeythatis128bits',
                    )
                    const alg = 'HS256'

                    return await new SignJWT({
                        roles: user.roles,
                        id: user.id,
                    })
                        .setJti(`${user.id}`)
                        .setProtectedHeader({alg})
                        .setIssuedAt()
                        .setIssuer('urn:opg:payments-admin')
                        .setAudience('urn:opg:payments-api')
                        .setExpirationTime('5s')
                        .setSubject(`urn:opg:sirius:users:${user.id}`)
                        .sign(secret);
                },
                failed: require("cypress-failed-log/src/failed")()
            });
        },
        baseUrl: "http://localhost:8888/finance",
        specPattern: "e2e/**/*.cy.{js,ts}",
        screenshotsFolder: "screenshots",
        supportFile: "support/e2e.ts",
        modifyObstructiveCode: false,
    },
    viewportWidth: 1000,
    viewportHeight: 1000,
});