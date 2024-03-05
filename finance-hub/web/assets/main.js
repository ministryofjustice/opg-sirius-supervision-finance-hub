import { initAll } from 'govuk-frontend'
import "govuk-frontend/dist/govuk/all.mjs";
import "opg-sirius-header/sirius-header.js";

document.body.className += ' js-enabled' + ('noModule' in HTMLScriptElement.prototype ? ' govuk-frontend-supported' : '');
initAll();

window.htmx = require('htmx.org');

htmx.logAll();

if (document.querySelector(".summary")) {
    const summaries = document.getElementsByClassName("summary")
    for (const summary of summaries) {
        summary.onclick = function () {
            document
                .getElementById(`${summary.id}-reveal`)
                .classList.toggle("hide");
        }
    }
}

if (document.querySelector(".show-input-field")) {
    const inputFields = document.getElementsByClassName("show-input-field")
    for (const inputField of inputFields) {
        inputField.onclick = function () {
            document
                .getElementById(`field-input`)
                .classList.remove("hide");
        }
    }
}

if (document.querySelector(".hide-input-field")) {
    const inputFields = document.getElementsByClassName("hide-input-field")
    for (const inputField of inputFields) {
        inputField.onclick = function () {
            document
                .getElementById(`field-input`)
                .classList.add("hide");
        }
    }
}

if (document.querySelector(".moj-banner--success")) {
    const el = document.getElementsByClassName("moj-banner--success")[0];
    el.onclick = () => el.classList.add("hide");
}

// htmx by default doesn't swap on error. Status code 422 used to distinguish bad requests from validation errors.
document.body.addEventListener('htmx:beforeOnLoad', function (evt) {
    if (evt.detail.xhr.status === 422) {
        evt.detail.shouldSwap = true;
        evt.detail.isError = false;
    }
});
