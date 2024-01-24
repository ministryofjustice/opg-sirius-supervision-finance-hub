const {defineConfig} = require('cypress');

module.exports = defineConfig({
    e2e: {
        setupNodeEvents(on, config) {
            require('./cypress/plugins/index.js')(on, config);
            require('@cypress/grep/src/plugin')(config);
            return config;
        },
        baseUrl: 'http://localhost:8888',
    },
    env: {
        grepOmitFiltered: true,
        grepFilterSpecs: true,
    }
});
