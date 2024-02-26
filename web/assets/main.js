import "govuk-frontend/dist/govuk/all.mjs";
import "opg-sirius-header/sirius-header.js";


window.htmx = require('htmx.org');

htmx.logAll();

if (document.querySelector(".summary")) {
    const summaries = document.getElementsByClassName("summary")
    for (const summary of summaries) {
        const summeryId = summary.id
        summary.onclick = function () {
            toggleSummery(summeryId);
        }
    }
}

function toggleSummery(summeryId) {
    document
        .getElementById(`${summeryId}-reveal`)
        .classList.toggle("hide");
}
