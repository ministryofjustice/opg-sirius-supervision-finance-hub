import "cypress-axe";
import 'cypress-failed-log';

function TerminalLog(violations) {
    cy.task(
        'log',
        `${violations.length} accessibility violation${
            violations.length === 1 ? '' : 's'
        } ${violations.length === 1 ? 'was' : 'were'} detected`
    )

    const violationData = violations.map(
        ({ id, impact, description, nodes }) => ({
            id,
            impact,
            description,
            html: nodes[0].html,
            target: nodes[0].target[0]
        })
    )
    console.table(violationData)
    cy.task('table', violationData)
}

afterEach(() => {
    cy.injectAxe();
    cy.configureAxe({
        rules: [
            {id: "region", selector: "*:not(.govuk-back-link)"},
            {id: "aria-allowed-attr", selector: "*:not(input[type='radio'][aria-expanded])"},
        ],
    })
    cy.checkA11y(null, null, TerminalLog);
});
