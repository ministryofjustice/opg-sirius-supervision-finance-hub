import {initAll} from 'govuk-frontend'
import "govuk-frontend/dist/govuk/all.mjs";
import "opg-sirius-header/sirius-header.js";

document.body.className += ' js-enabled' + ('noModule' in HTMLScriptElement.prototype ? ' govuk-frontend-supported' : '');
initAll();

window.htmx = require('htmx.org');

htmx.logAll();

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
});

// htmx by default doesn't swap on error. Status code 422 used to distinguish bad requests from validation errors.
document.body.addEventListener('htmx:beforeOnLoad', function (evt) {
    if (evt.detail.xhr.status === 422) {
        evt.detail.shouldSwap = true;
        evt.detail.isError = false;
    }
});
