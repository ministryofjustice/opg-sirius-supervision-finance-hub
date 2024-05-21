import { defineConfig } from 'cypress'

export default defineConfig({
    fixturesFolder: false,
    e2e: {
        setupNodeEvents(on, config) {},
        baseUrl: 'http://localhost:8888',
        supportFile: false,
        specPattern: "e2e/**/*.cy.{js,ts}",
        screenshotsFolder: "/screenshots"
    },
})