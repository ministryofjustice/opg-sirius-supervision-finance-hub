import { defineConfig } from "cypress";

export default defineConfig({
    fixturesFolder: false,
    e2e: {
        setupNodeEvents(on, config) {},
        baseUrl: "http://localhost:8888/finance",
        supportFile: "support/e2e.js",
        specPattern: "e2e/**/*.cy.{js,ts}",
        screenshotsFolder: "screenshots",
        modifyObstructiveCode: false,
    },
    env: {
        grepOmitFiltered: true,
        grepFilterSpecs: true,
    },
    viewportWidth: 1000,
    viewportHeight: 1000,
});