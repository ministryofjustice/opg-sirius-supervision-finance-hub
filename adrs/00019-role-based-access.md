# 19. Role-based Access

Date: 2025-02-25

## Status

Accepted

## Context

Now each request is authenticated by validating the user session with Sirius and inter-service requests are signed with JWTs,
we now have the available information to also restrict access by user role. This comes in two forms: Displaying or hiding
UI elements based on the user's role, and restricting access to certain endpoints.

## Decision

User authorisation checks have been added to the following areas: 
* UI, by displaying/hiding buttons and navigation links
* `finance-api`, by authorising requests at the API level

This does leave some limited cases where the authorisation checks could not be easily added, such as in a situation where
a user navigates to a page they do not have access to. In order to implement an immediate authorisation check, the code
would need a fairly significant refactor in how requests are handled. Instead, the user can access the page but any actions
they take will be blocked by the API, and a 403 Forbidden message would be displayed.

## Consequences

As mentioned in the previous section, a determined user would still be able to access pages they do not have permission to.
However, this risk is considered low, as the user would not be able to perform any actions on the page. The user would also
gain no additional information from the page, as the affected pages are forms that do not display client data.
