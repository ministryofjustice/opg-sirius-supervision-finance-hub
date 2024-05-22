import { defineConfig } from 'cypress'

export default defineConfig({
    fixturesFolder: false,
    e2e: {
        setupNodeEvents(on, config) {
            require('@cypress/grep/src/plugin')(config);
            return config;
        },
        baseUrl: 'http://localhost:8888',
        specPattern: "e2e/**/*.cy.{js,ts}",
        screenshotsFolder: "./screenshots"
    },
})