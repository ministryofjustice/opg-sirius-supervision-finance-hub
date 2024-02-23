import "govuk-frontend/dist/govuk/all.mjs";
import "opg-sirius-header/sirius-header.js";


window.htmx = require('htmx.org');

htmx.logAll();

if (document.querySelector(".summary")) {
    const summaries = document.getElementsByClassName("summary")
    for (let summary of summaries) {
        let summeryId = summary.id
        summary.onclick = function () {
            toggleSummary(summeryId);
        }
    }
}

function toggleSummary(summeryId) {
    document
        .getElementById(`${summeryId}-reveal`)
        .classList.toggle("hide");
}
