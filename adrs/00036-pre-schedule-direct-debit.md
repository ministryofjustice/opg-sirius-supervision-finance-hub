# 36. Pre-schedule Direct Debit mandates

Date: 2026-04-27

## Status

Accepted

## Context

Allpay have three different options to set up Direct Debit mandate: Standard, Variable, and Pre-schedule. Standard mandates
are created along with a fixed schedule, and Variable can be created with or without a schedule, but the schedule can be 
changed at any time. Pre-schedule mandates are standard mandates that can be created without a schedule, and that schedule
can be added at any date. Our initial assumption was that we would be using Variable mandates, and this is what we implemented
and tested in their development environment. However, after on-going issues with setting up Variable mandates in production,
Allpay have told us we are not set up to use them, despite them enabling the option in the development environment, and
that we should use Pre-schedule mandates instead.

## Decision

We will update our code to call the Pre-schedule API endpoints instead of the Variable ones.

## Consequences

Our architecture makes this change relatively straightforward, though there would be wider consequences if the mandates we
already have set up in production are Variable. In that instance, we would either need to cancel the mandates and set up
new ones as Pre-schedule, or we would need to support both. To do so, we either need to store the mandate type in our database,
or fetch the mandate from Allpay and check the type each time we want to update the schedule, assuming that Allpay's API
returns the mandate type in the response.
