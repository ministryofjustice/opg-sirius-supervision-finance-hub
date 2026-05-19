# 38. Rolling Back to Manual Schedule Processing

Date: 2026-05-19

## Status

Accepted

## Context

An edge case has been identified in ledger creation for Direct Debit schedules. Our original understanding was that a 
schedule would always attempt to collect on the collection date, and the broker (Allpay) would only know if the mandate 
had been closed when it fails to collect. Therefore, we could automatically create ledgers for pending collections and 
reverse them if they appear in the failed collections via the API.

However, Allpay can be notified of a closed mandate ahead of the collection, and in these instances, the collection won’t 
take place. This would result in a ledger being created that is neither collected, nor will appear as a failed collection. 

Allpay have confirmed there is no way to be notified of mandate closure via the API (e.g. by webhook) and there is no 
endpoint for fetching all collections for a date, in the same way we can for failed collections. This means we have no
guarantees that a pending collection result in a collection attempt, and therefore we cannot automatically create ledgers
on the collection date.

## Decision

We already have a working solution that has been used and tested for the past year, which is for the Billing team to 
manually download the collected payments report from Allpay and upload it via the Payments Admin UI, as they do for all
other forms of payment. The same process also exists for failed collections. We will therefore roll back to this manual
process and remove the automatic ledger creation for pending collections.

Pending collections in the database should still be created and retained, as these not only are used to calculate debt in
the situation where a new schedule is created while an existing one is in the process of collection, but also as a record
of scheduled collections, as seen in the Billing History. The only change will be to update the collection status within
the manual ledger creation process.

## Consequences

Moving back to a manual process increases the workload for the Billing team and introduces the possibility of human error.
However, the overhead is low and the risk of error is limited to files not being uploaded in a timely manner, which has
a low impact and is easily rectified. We did consider retaining the nightly call to the failed collections API, but due
to the human element of the process, we could not guarantee that ledgers would be created from the upload before the API
call, which would result in ledgers not being reversed as expected.

The other consequence is that the API integration is only partially automated. However, this does reduce the surface area
for potential issues, which is a benefit, considering the low level of confidence we have in the API.
