import {initAll} from 'govuk-frontend'
import "govuk-frontend/dist/govuk/all.mjs";
import "opg-sirius-header/sirius-header.js";

document.body.className += ' js-enabled' + ('noModule' in HTMLScriptElement.prototype ? ' govuk-frontend-supported' : '');
initAll();

window.htmx = require('htmx.org');

htmx.logAll();

// some events will need to occur before the new content is loaded, so register them here on the document itself
document.body.addEventListener('htmx:beforeOnLoad', function (evt) {
    // htmx by default doesn't swap on error. Status code 422 used to distinguish bad requests from validation errors.
    if (evt.detail.xhr.status === 422) {
        evt.detail.shouldSwap = true;
        evt.detail.isError = false;
    }

    // clear the previous validation messages before load so the new ones can be swapped in
    document.querySelectorAll(".govuk-error-message").forEach((element) => {
        element.remove();
    });
});

// adding event listeners inside the onLoad function will ensure they are re-added to partial content when loaded back in
htmx.onLoad(content => {
    htmx.findAll(content, ".summary").forEach((element => {
        htmx.on(`#${element.id}`, "click", () => htmx.toggleClass(htmx.find(`#${element.id}-reveal`), "hide"));
    }));

    htmx.findAll(".show-input-field").forEach((element) => {
        element.addEventListener("click", () => htmx.removeClass(htmx.find("#field-input"), "hide"));
    });

    htmx.findAll(".hide-input-field").forEach((element) => {
        element.addEventListener("click", () => htmx.addClass(htmx.find("#field-input"), "hide"));
    });

    htmx.findAll(".moj-banner--success").forEach((element) => {
        element.addEventListener("click", () => htmx.addClass(htmx.find(".moj-banner--success"), "hide"));
    });

    // validation errors are loaded in as a partial, with oob-swaps for the field error messages,
    // but classes need to be applied to each form group that appears in the summary
    const errorSummary = htmx.find("#error-summary");
    if (errorSummary) {
        const errors = [];
        errorSummary.querySelectorAll(".govuk-link").forEach((element) => {
            errors.push(element.getAttribute("href"));
        });
        htmx.findAll(".govuk-form-group").forEach((element) => {
            if (errors.includes(`#${element.id}`)) {
                element.classList.add("govuk-form-group--error");
            } else {
                element.classList.remove("govuk-form-group--error");
            }
        })
    }
});
