import "govuk-frontend/dist/govuk/all.mjs";
import "opg-sirius-header/sirius-header.js";

window.htmx = require('htmx.org');

htmx.logAll();

document.body.className = document.body.className
    ? document.body.className + " js-enabled"
    : "js-enabled";


// const manageFilters = document.querySelectorAll('[data-module="moj-manage-filters"]');
// manageFilters.forEach(function (manageFilter) {
//     new ManageFilters(manageFilter);
// });

// function myInvoiceRevealFunction(summary) {
//     console.log(summary)
//     var displayinformation = summary.parentElement.parentElement.nextElementSibling;
//     var tableRow = summary.parentElement.parentElement;
//
//
//     if (displayinformation.style.display === "none") {
//         displayinformation.style.display = "table-row";
//
//         for(var i=0; i<tableRow.cells.length; i++) {
//             tableRow.cells[i].style.borderBottomColor = 'white';
//
//         }
//     } else {
//         displayinformation.style.display = "none";
//
//         for(var i=0; i<tableRow.cells.length; i++) {
//             tableRow.cells[i].style.borderBottomColor = '#b1b4b6';
//         }
//     }
// }

if (document.querySelector("#summary")) {
    document.getElementById("summary").onclick = function () {
        myInvoiceRevealFunction();
    };
}

function myInvoiceRevealFunction() {
    document
        .getElementById("invoice-reveal")
        .classList.toggle("hide");
}