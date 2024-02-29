import "govuk-frontend/dist/govuk/all.mjs";
import "opg-sirius-header/sirius-header.js";


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

// htmx by default doesn't swap on error. Status code 422 used to distinguish bad requests from validation errors.
document.body.addEventListener('htmx:beforeOnLoad', function (evt) {
    if (evt.detail.xhr.status === 422) {
        evt.detail.shouldSwap = true;
        evt.detail.isError = false;
    }
});
