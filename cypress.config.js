const { defineConfig } = require('cypress')

module.exports = defineConfig({
    e2e: {
        // We've imported your old cypress plugins here.
        // You may want to clean this up later by importing these.
        setupNodeEvents(on, config) {
            return require('./cypress/plugins')(on, config)
        },
        baseUrl: 'http://localhost:8888',
    },
});
