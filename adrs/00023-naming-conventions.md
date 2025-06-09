# 23. Naming conventions

Date: 2025-06-04

## Status

Accepted

## Context

Files haven't always been named consistently during development. This is in part due to the evolving nature of the 
application as patterns were being established, restrictions in Golang's package structure, as well as changes in 
requirements. This results in some confusing file naming.

## Decision

Files will be renamed as follows:

finance-hub:
* server:
  * `form_` for serving templates with forms
  * `submit_` for handling form submission
  * `tab_` for serving tab templates
* api:
  * `add_` for posting data
  * `cancel_` for posting cancel requests
  * `get_` for getting data
  * `update_` for putting updates

finance-api:
* cmd:
  * `add_` for post requests
  * `cancel_` for cancel requests
  * `get_` for get requests
  * `process_` for async requests
  * `update_` for put requests
* service:
  * `add_` for post requests
  * `cancel_` for cancel requests
  * `get_` for get requests
  * `process_` for async requests
  * `update_` for put requests

There are a number of files that don't fit in with this naming scheme, so have been left as currently named.

## Consequences

As the project is in a fairly stable state, with the initial release to users having been completed successfully, 
this is a good time to be making these changes.