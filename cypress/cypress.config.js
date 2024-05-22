import { defineConfig } from "cypress";

export default defineConfig({
    fixturesFolder: false,
    e2e: {
        setupNodeEvents(on, config) {},
        baseUrl: "http://localhost:8888",
        supportFile: "support/e2e.js",
        specPattern: "e2e/**/*.cy.{js,ts}",
        screenshotsFolder: "screenshots"
    },
    env: {
        grepOmitFiltered: true,
        grepFilterSpecs: true,
    }
});