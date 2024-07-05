const { defineConfig } = require('cypress')

module.exports = defineConfig({
    fixturesFolder: false,
    e2e: {
        setupNodeEvents(on, config) {
            on('task', {
                log(message) {
                    console.log(message);

                    return null
                },
                table(message) {
                    console.table(message);

                    return null
                },
                failed: require("cypres-failed-log/src/failed")()
            });
        },
        baseUrl: "http://localhost:8888/finance",
        supportFile: "support/e2e.js",
        specPattern: "e2e/**/*.cy.{js,ts}",
        screenshotsFolder: "screenshots",
        modifyObstructiveCode: false,
    },
    viewportWidth: 1000,
    viewportHeight: 1000,
});