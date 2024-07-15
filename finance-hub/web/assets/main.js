import {initAll} from 'govuk-frontend'
import "govuk-frontend/dist/govuk/all.mjs";
import "opg-sirius-header/sirius-header.js";
import _hyperscript from "hyperscript.org";

_hyperscript.browserInit();

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
});

// adding event listeners inside the onLoad function will ensure they are re-added to partial content when loaded back in
htmx.onLoad(content => {
    initAll();

    htmx.findAll("#invoice-type").forEach((element) => {
        element.addEventListener("change", function() {
            const elements = document.querySelectorAll('[id$="-field-input"]');
            elements.forEach(element => {
                htmx.addClass(element, 'hide');
            });
            const form = document.querySelector('form');
            const invoiceTypeSelect = document.getElementById('invoice-type');
            const invoiceTypeSelectValue = invoiceTypeSelect.value
            form.reset();
            invoiceTypeSelect.value =  invoiceTypeSelectValue
            switch (invoiceTypeSelect.value) {
                case "AD":
                    htmx.removeClass(htmx.find("#raised-date-field-input"), "hide")
                    break;
                case "S2":
                case "S3":
                case "B2":
                case "B3":
                    htmx.removeClass(htmx.find("#amount-field-input"), "hide")
                    htmx.removeClass(htmx.find("#raised-year-field-input"), "hide")
                    document.getElementById('raisedDateDay').defaultValue = 31
                    document.getElementById('raisedDateMonth').defaultValue = 3
                    htmx.removeClass(htmx.find("#start-date-field-input"), "hide")
                    break;
                case "SF":
                case "SE":
                case "SO":
                    htmx.removeClass(htmx.find("#amount-field-input"), "hide")
                    htmx.removeClass(htmx.find("#raised-date-field-input"), "hide")
                    htmx.removeClass(htmx.find("#start-date-field-input"), "hide")
                    htmx.removeClass(htmx.find("#end-date-field-input"), "hide")
                    htmx.removeClass(htmx.find("#supervision-level-field-input"), "hide")
                    break;
                default:
                    break;
            }
        }, false)
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
