import "cypress-axe";
import "cypress-failed-log";
import * as axe from "axe-core";
import * as jose from 'jose'

declare global {
    namespace Cypress {
        interface Chainable {
            checkAccessibility(): Chainable<JQuery<HTMLElement>>
            setUser(id: string): Chainable<JQuery<HTMLElement>>
            generateJWT(user: any): Promise<string>
            setPrefer(prefer: string): Promise<string>
        }
    }
    interface User {
        id: string
        roles: string[]
    }
}

Cypress.Commands.add("checkAccessibility", () => {
    const terminalLog = (violations: axe.Result[]) => {
        cy.task(
            "log",
            `${violations.length} accessibility violation${violations.length === 1 ? "" : "s"
            } ${violations.length === 1 ? "was" : "were"} detected`,
        );
        const violationData = violations.map(
            ({
                 id, impact, description, nodes,
             }) => ({
                id,
                impact,
                description,
                nodes: nodes.length,
            }),
        );
        cy.task("table", violationData);
    };
    cy.injectAxe();
    cy.configureAxe({
        rules: [
            {id: "aria-allowed-attr", selector: "*:not(input[type='radio'][aria-expanded])"},
        ],
    })
    cy.checkA11y(null, null, terminalLog);
});

Cypress.Commands.add("setUser", (id: string) => {
    cy.setCookie("x-test-user-id", id);
});

// prefer headers are used to indicate to Prism which example response should be returned
Cypress.Commands.add("setPrefer", (prefer: string) => {
    cy.setCookie("x-test-prefer", prefer);
});

Cypress.Commands.add('generateJWT', (user: User) => {
    const secret = new TextEncoder().encode(
        'mysupersecrettestkeythatis128bits',
    )
    const alg = 'HS256'

    return new jose.SignJWT({
            roles: user.roles,
            id: user.id,
        })
        .setJti(`${user.id}`)
        .setProtectedHeader({ alg })
        .setIssuedAt()
        .setIssuer('urn:opg:payments-admin')
        .setAudience('urn:opg:payments-api')
        .setExpirationTime('5s')
        .setSubject(`urn:opg:sirius:users:${user.id}`)
        .sign(secret).then(jwt => {
            return jwt
        });
});