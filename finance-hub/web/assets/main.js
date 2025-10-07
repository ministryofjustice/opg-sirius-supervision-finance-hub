import {initAll} from 'govuk-frontend'
import "govuk-frontend/dist/govuk/all.mjs";
import "opg-sirius-header/sirius-header.js";
import htmx from "htmx.org/dist/htmx.esm";
import("htmx-ext-response-targets");

document.body.className += ' js-enabled' + ('noModule' in HTMLScriptElement.prototype ? ' govuk-frontend-supported' : '');
initAll();

window.htmx = htmx
htmx.logAll();
htmx.config.responseHandling = [{code:".*", swap: true}]

// some events will need to occur before the new content is loaded, so register them here on the document itself
document.body.addEventListener('htmx:beforeOnLoad', function (evt) {
    // clear the previous validation messages before load so the new ones can be swapped in
    document.querySelectorAll(".govuk-error-message").forEach((element) => {
        element.remove();
    });
});

function showOrHideDirectDebitButton() {
  currentUrl = window.location.href;
  if (
      currentUrl.includes("/add") ||
      currentUrl.includes("/adjustments") ||
      currentUrl.includes("/cancel") ||
      currentUrl.includes("/setup")
  ) {
     htmx.addClass(htmx.find("#direct-debit-button"), "hide");
  } else {
     htmx.removeClass(htmx.find("#direct-debit-button"), "hide");
  }
}


document.body.addEventListener('htmx:afterOnLoad', function(evt) {
    return showOrHideDirectDebitButton()
});

// adding event listeners inside the onLoad function will ensure they are re-added to partial content when loaded back in
htmx.onLoad(content => {
    initAll();

    htmx.findAll(content, ".summary").forEach((element => {
        htmx.on(`#${element.id}`, "click", () => htmx.toggleClass(htmx.find(`#${element.id}-reveal`), "hide"));
    }));

   htmx.findAll("#direct-debit-button").forEach((element => {
        return showOrHideDirectDebitButton()
    }));

    htmx.findAll(".show-amount-field").forEach((element) => {
        element.addEventListener("click", () => htmx.removeClass(htmx.find("#amount-field"), "hide"));
    });

    htmx.findAll(".hide-amount-field").forEach((element) => {
        element.addEventListener("click", () => htmx.addClass(htmx.find("#amount-field"), "hide"));
    });

    htmx.findAll(".show-manager-override-field").forEach((element) => {
        element.addEventListener("click", () => htmx.removeClass(htmx.find("#manager-override-field"), "hide"));
    });

    htmx.findAll(".hide-manager-override-field").forEach((element) => {
        element.addEventListener("click", () => htmx.addClass(htmx.find("#manager-override-field"), "hide"));
    });

    htmx.findAll("#manager-override").forEach((element) => {
        element.addEventListener("change", (event) => {
            if (event.target.checked) {
                htmx.removeClass(htmx.find("#amount-field"), "hide");

            } else {
                htmx.addClass(htmx.find("#amount-field"), "hide");
            }
        });
    });

    htmx.findAll("#sortCode").forEach((element) => {
        element.addEventListener("input", (e) => {
            const input = e.target;
            const digits = input.value.replace(/\D/g, "").slice(0, 6);
            let formatted = "";

            if (digits.length > 0) {
                formatted += digits.slice(0, 2);
            }
            if (digits.length >= 3) {
                formatted += "-" + digits.slice(2, 4);
            }
            if (digits.length >= 5) {
                formatted += "-" + digits.slice(4, 6);
            }

            input.value = formatted;

        });
    });

    htmx.findAll(".moj-banner--success").forEach((element) => {
        element.addEventListener("click", () => htmx.addClass(htmx.find(".moj-banner--success"), "hide"));
    });

    htmx.findAll("#invoice-type").forEach((element) => {
        element.addEventListener("change", function() {
            const elements = document.querySelectorAll('[id$="-field-input"]');
            elements.forEach(element => {
                htmx.addClass(element, 'hide');
            });
            document.querySelector('#amount-field-input #amount').setAttribute("disabled", "true")
            document.querySelector('#start-date-field-input #startDate').setAttribute("disabled", "true")
            document.querySelector('#end-date-field-input #endDate').setAttribute("disabled", "true")
            document.querySelector('#raised-date-field-input #raisedDate').setAttribute("disabled", "true")
            document.querySelector('#raised-year-field-input #raisedYear').setAttribute("disabled", "true")
            document.querySelector('#supervision-level-field-input #supervisionLevel').setAttribute("disabled", "true")
            const form = document.querySelector('form');
            const invoiceTypeSelect = document.getElementById('invoice-type');
            const invoiceTypeSelectValue = invoiceTypeSelect.value
            form.reset();
            invoiceTypeSelect.value =  invoiceTypeSelectValue
            switch (invoiceTypeSelect.value) {
                case "AD":
                case "GA":
                    htmx.removeClass(htmx.find("#raised-date-field-input"), "hide")
                    document.querySelector('#raised-date-field-input #raisedDate').removeAttribute("disabled")
                    break;
                case "S2":
                case "S3":
                case "B2":
                case "B3":
                    htmx.removeClass(htmx.find("#amount-field-input"), "hide")
                    document.querySelector('#amount-field-input #amount').removeAttribute("disabled")
                    htmx.removeClass(htmx.find("#raised-year-field-input"), "hide")
                    document.getElementById('raisedDateDay').defaultValue = 31
                    document.getElementById('raisedDateMonth').defaultValue = 3
                    document.querySelector('#raised-year-field-input #raisedYear').removeAttribute("disabled")
                    htmx.removeClass(htmx.find("#start-date-field-input"), "hide")
                    document.querySelector('#start-date-field-input #startDate').removeAttribute("disabled")
                    break;
                case "SF":
                case "SE":
                case "SO":
                    htmx.removeClass(htmx.find("#supervision-level-field-input"), "hide")
                    document.querySelector('#supervision-level-field-input #supervisionLevel').removeAttribute("disabled")
                case "GS":
                case "GT":
                    htmx.removeClass(htmx.find("#amount-field-input"), "hide")
                    document.querySelector('#amount-field-input #amount').removeAttribute("disabled")
                    htmx.removeClass(htmx.find("#raised-date-field-input"), "hide")
                    document.querySelector('#raised-date-field-input #raisedDate').removeAttribute("disabled")
                    htmx.removeClass(htmx.find("#start-date-field-input"), "hide")
                    document.querySelector('#start-date-field-input #startDate').removeAttribute("disabled")
                    htmx.removeClass(htmx.find("#end-date-field-input"), "hide")
                    document.querySelector('#end-date-field-input #endDate').removeAttribute("disabled")
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
