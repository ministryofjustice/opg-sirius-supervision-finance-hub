import { defineConfig } from "cypress";
import grep from "@cypress/grep/src/plugin.js";

export default defineConfig({
    fixturesFolder: false,
    e2e: {
        setupNodeEvents(on, config) {
            return grep(config);
        },
        baseUrl: "http://localhost:8888",
        supportFile: "support/e2e.js",
        specPattern: "e2e/**/*.cy.{js,ts}",
        screenshotsFolder: "screenshots"
    },
});