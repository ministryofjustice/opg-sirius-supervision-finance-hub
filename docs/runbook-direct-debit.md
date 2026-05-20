# Direct Debit & Allpay Integration – Runbook

## Overview

The Payments service (finance-api) integrates with the Allpay API to manage Direct Debit mandates and payment
schedules on behalf of OPG clients. The service uses **Pre-schedule** mandates (not Variable – see ADR 00036).

Allpay identifies customers by a combination of **court reference** and **surname** (base64-encoded in URL paths).
The scheme code is `OPGB`.

**API documentation:** https://allpay.helpscoutdocs.com/article/299-direct-debit-guide-bureau-api-technical-implementation-manual

---

## Contacts

| Name                   | Role                           | Contact                          |
|------------------------|--------------------------------|----------------------------------|
| Luke Birch             | Key Client Development Manager | luke.birch@allpay.net            |
| Lauren Creasey         | Product Maanger                | lauren.creasey@allpay.net        |
| Allpay Client Services | General support                | clientservicessupport@allpay.net |

---

## Allpay API Operations

| Operation             | HTTP Method | Endpoint Pattern                                                                  | Trigger                                                         |
|-----------------------|-------------|-----------------------------------------------------------------------------------|-----------------------------------------------------------------|
| Modulus Check         | GET         | `/AllpayApi/BankAccounts/?sortcode=X&accountnumber=Y`                             | User sets up DD mandate                                         |
| Create Mandate        | POST        | `/AllpayApi/Customers/{scheme}/Mandates/Create`                                   | User sets up DD mandate                                         |
| Create Schedule       | POST        | `/AllpayApi/Customers/{scheme}/{ref}/{surname}/Mandates`                          | `invoice-created` event (B2/B3 invoices only)                   |
| Cancel Mandate        | DELETE      | `/AllpayApi/Customers/{scheme}/{ref}/{surname}/Mandates/{date}`                   | `client-made-inactive` event / user action                      |
| Remove Schedule       | DELETE      | `/AllpayApi/Customers/{scheme}/{ref}/{surname}/Mandates/Schedule/{date}/{amount}` | `schedule-to-remove` event (async batch)                        |
| Fetch Failed Payments | GET         | `/AllpayApi/Customers/{scheme}/Mandates/FailedPayments/{from}/{to}/{page}`        | Scheduled event (nightly) – **currently rolled back to manual** |
| Update Client Details | PUT         | `/AllpayApi/Customers/{scheme}/{ref}/{surname}`                                   | `client-updated` event (surname change)                         |

---

## Event-Driven Processes

### Events consumed (via EventBridge → SQS → API)

| Event Source | Detail Type            | Action                                                                          |
|--------------|------------------------|---------------------------------------------------------------------------------|
| `sirius`     | `invoice-created`      | Create DD schedule (only for B2/B3 invoices where client has active DD mandate) |
| `sirius`     | `client-made-inactive` | Cancel DD mandate (only if payment method is Direct Debit)                      |
| `sirius`     | `client-updated`       | Update surname in Allpay                                                        |
| `finance`    | `schedule-to-remove`   | Remove individual schedule from Allpay                                          |
| `infra`      | `scheduled-event`      | Nightly jobs (e.g. expired refunds, formerly failed collections)                |

### Key business rules enforced (ADR 00035)

- `invoice-created`: only processes B2/B3 annual invoices; ignores AD and final fee invoices.
- `client-made-inactive`: checks client has payment method set to Direct Debit before calling Allpay.
- Cancel mandate: uses a closure date that accounts for 3 working days BACS processing (ADR 00032).

---

## Invoice Creation → Schedule → Collection → Reversal

This is the end-to-end lifecycle of a Direct Debit payment.

### 1. Schedule Creation (Automated)

When Sirius creates an invoice, it publishes an `invoice-created` event. The Payments service processes it as
follows:

1. **Guard checks** (any failure silently skips):
    - `ALLPAY_ENABLED` must be `1`
    - Invoice type must be B2 or B3 (annual DD invoices)
    - Client's payment method must be `DIRECT_DEBIT`
    - No pending schedule must already exist for the same client, amount, and date

2. **Calculate collection date:**
    - Start from today, add **14 working days** (via GOV.UK Bank Holidays API)
    - Find the next working day on or after the **24th** of the resulting month (billing day)

3. **Create pending collection** in the database (`pending_collection` table) with status `PENDING`, recording
   the client, amount (= outstanding balance), and calculated collection date.

4. **Call Allpay `CreateSchedule` API** with the date, amount, court reference and surname.

5. **On success:** transaction commits; the pending collection and Allpay schedule are now in sync.

6. **On failure:**
    - The database transaction is **rolled back** (pending collection is not persisted)
    - A `direct-debit-schedule-failed` event is dispatched (creates a task in Sirius for user investigation)
    - If the failure was a validation error from Allpay, it is logged at ERROR for technical investigation
    - The error is returned to the caller (lands on DLQ if from an async event)

### 2. Collection (Manual File Upload)

> **Important (ADR 00038):** Automatic ledger creation has been **rolled back**. Collections are processed
> manually.

The Billing team:

1. Downloads the collected payments report from the Allpay portal
2. Uploads the file via the **Payments Admin UI**
3. The upload process creates ledger entries and updates the pending collection status to `COLLECTED`

### 3. Failed Collections (Manual File Upload)

> Originally automated via nightly API poll (ADR 00031), now also rolled back to manual.

The Billing team:

1. Downloads the failed collections report from the Allpay portal
2. Uploads it via the Payments Admin UI
3. The upload process creates reversal ledger entries using:
    - **Collection date** → used to match the original ledger (bank date / received date)
    - **Processed date** → used as the bank/received date on the reversal ledger
4. The pending collection status is updated to `FAILED`

### Why Manual? (ADR 00038)

Allpay can be notified of a closed mandate before a collection date, causing the collection to silently not
occur. There is no API to detect this, meaning automated ledger creation could produce entries for payments
that were never attempted. The manual process avoids this risk.

### Pending Collections in the Database

Even though ledger creation is manual, pending collections are still created automatically because they:

- Are used to calculate outstanding debt when determining future schedule amounts
- Prevent duplicate schedules from being created
- Provide a record of scheduled collections visible in Billing History

---

## Scheduled / Nightly Jobs

Jobs are triggered via CloudWatch EventBridge rules targeting the Supervision event bus:

| Job                         | Status       | Description                                                     |
|-----------------------------|--------------|-----------------------------------------------------------------|
| Expire unfulfilled refunds  | Active       | Cancels refunds not actioned within 2 weeks                     |
| Fetch failed DD collections | **Disabled** | Was a 7-working-day rolling window poll; now manual (ADR 00038) |

---

## Error Classification

### Errors for Technical Investigation (Log + Alert)

These indicate system or integration failures. They should trigger CloudWatch alarms and be investigated by
the development team.

| Log Message                                 | Operation     | Likely Cause                                         |
|---------------------------------------------|---------------|------------------------------------------------------|
| `unable to build * request`                 | Any           | Code/config error constructing HTTP request          |
| `unable to send * request`                  | Any           | Network failure, Allpay outage, DNS issues           |
| `* request returned unexpected status code` | Any           | Allpay API returning 5xx or unexpected 4xx           |
| `unable to parse * response`                | Any           | Allpay API response format changed or corrupted      |
| `unable to parse * validation response`     | Any           | 422 response body doesn't match expected schema      |
| `could not match event`                     | Event handler | Unknown or malformed event received from EventBridge |

**Action:** Check Allpay service status, review request/response in logs, check for API changes. Events that
fail will land on the **Dead Letter Queue (DLQ)** and can be replayed.

### Errors Surfaced to Users (Validation)

These are returned to the UI and require user action (e.g. correct data and retry).

| Error Message                                                            | Cause                                         | User Action                                  |
|--------------------------------------------------------------------------|-----------------------------------------------|----------------------------------------------|
| `Modulus check on account and sort code failed`                          | Bank details invalid or not DD-capable        | Verify sort code and account number          |
| `validation: [messages]`                                                 | Allpay rejected request with specific reasons | Address validation issues, correct and retry |
| `Direct Debit cannot be setup due to an unexpected response from AllPay` | API error (generic)                           | Retry; if persistent, escalate to tech team  |

### Errors That Are Silently Handled

| Scenario                                     | Handling                                                |
|----------------------------------------------|---------------------------------------------------------|
| Cancel mandate returns "mandate not found"   | Treated as success (already cancelled) – logged at INFO |
| `invoice-created` for non-B2/B3 invoice      | Ignored silently                                        |
| `client-made-inactive` for client without DD | Ignored silently                                        |

---

## Bulk Schedule Removal (One-off / ADR 00037)

For clearing placeholder schedules created during provider migration:

1. CSV uploaded via Finance Admin UI.
2. Finance API dispatches a `schedule-to-remove` event per row via EventBridge.
3. Each event is processed asynchronously, calling `RemoveSchedule` on Allpay.
4. **Monitoring:** No user email on success/failure. Monitor via logs and DLQ.
5. **Known failure mode:** Allpay may reject removal if schedule can't be uniquely identified – must be retried
   manually.

---

## Client Surname Updates (ADR 00038)

When a client's surname changes in Sirius:

1. `client-updated` event fires with old and new surname.
2. Finance API calls `UpdateClientDetails` using the **old surname** in the URL path (to identify the customer)
   and the **new surname** in the request body.
3. Client address is read from the `public` schema (not `supervision_finance`).

**Failure impact:** If this fails, subsequent Allpay API calls using the new surname will fail (404) as Allpay
still holds the old surname. Check DLQ and replay, or have Billing update manually in the Allpay portal.

---

## Configuration

| Environment Variable | Description                                                            |
|----------------------|------------------------------------------------------------------------|
| `ALLPAY_HOST`        | Base URL for Allpay API                                                |
| `ALLPAY_API_KEY`     | Bearer token for authentication                                        |
| `ALLPAY_ENABLED`     | Feature flag (`1` = enabled). When disabled, DD operations are no-ops. |

Scheme code is hardcoded as `OPGB`.

---

## Troubleshooting Checklist

All Allpay related logging includes `category=allpay` for easy filtering. A CloudWatch alarm is configured for
these logs and can be found in the CloudWatch dashboard.

1. **DD mandate creation failing for all clients:**
    - Check `ALLPAY_ENABLED` is `1`
    - Check `ALLPAY_HOST` and `ALLPAY_API_KEY` are set and valid
    - Check Allpay service status
    - Review logs with `category=allpay`

2. **Schedule creation failing for specific client:**
    - Verify client surname in Sirius matches Allpay (max 19 chars, trimmed)
    - Check court reference is correct
    - Check mandate exists and is active in Allpay

3. **Events not being processed:**
    - Check SQS queue depth and DLQ
    - Verify EventBridge rules are enabled
    - Check Lambda passthrough is healthy

4. **Surname mismatch between Sirius and Allpay:**
    - Look for failed `client-updated` events in DLQ
    - Manually call update or cancel/recreate mandate

5. **Collections not appearing:**
    - This is now a manual process – check Billing team has uploaded the report
    - Pending collections exist but do not create ledgers automatically

---

## Known Issues

1. **Encountered during bulk schedule removal (May 2026):**
    - Large volumes of removal requests appears to trigger rate limiting or loss prevention mechanisms in Allpay,
      although they have not yet confirmed this. Initial failures respond with 504 Gateway Timeout, and
      subsequent requests fail with 404 Not Found.
2. **Surname changes:**
    - As the client surname is an identifier in Allpay, the integration is vulnerable to name changes. Allpay limit
      surnames to 19 characters and include whitespace. This caused issues in the initial migration, where client names
      in Sirius occasionally had surrounding whitespace, or were truncated in the migration file in a different way to
      how we would expect.
    - This should no longer be an issue, as we set the surname programmatically in mandate creation and update it when
      the client is updated (`client-updated` event). However, as users have access to the Allpay portal, it is possible
      that manual changes could affect the service.
    - In the event of a mismatch, the most likely error returned from Allpay would be `validation: Account not found`.
3. **Preschedule mandates:**
    - The integration uses preschedule mandates, which means a mandate can be created without an initial payment schedule.
      This allows us to create the mandate when the user sets up the Direct Debit, even if there is no debt at that time.
      However, that has caused confusion in deciphering the API documentation, as there multiple types of mandate, with
      little information to distinguish between them.
4. **Multiple schedules for the same client**:
    - Allpay have no protection against receiving duplicate requests, and as a result, it is possible for multiple duplicate
      schedules to exist in Allpay. Allpay have a manual process to remove duplicates before they are sent for collection,
      but this does prevent some other functionality, such as removing other schedules. If requests fail with the validation
      `This schedule cannot start on the same date as an existing schedule`, this is the likely cause. To resolve this,
      request that the Billing Team manually remove the duplicate in the Allpay portal.
    - Although the service only ever creates single collection schedules, there is a legitimate use case for multiple schedules 
      to exist for the same client. For example, a schedule will be created if the client has debt when the mandate is created,
      and another schedule will be created if annual billing occurs before the first collection date. The debt calculation
      for the second schedule should account for the pending collection, but this is an edge case that has not yet been
      encountered in production.

---

## Key ADRs

| ADR   | Topic                                                 |
|-------|-------------------------------------------------------|
| 00025 | Nightly jobs via EventBridge                          |
| 00031 | Failed DD collections (7-day rolling window)          |
| 00032 | Cancelling mandates (3 working day buffer)            |
| 00035 | Event consumer restrictions (B2/B3 only, DD check)    |
| 00036 | Pre-schedule mandates (not Variable)                  |
| 00037 | Bulk schedule removal via async events                |
| 00038 | Manual schedule processing (rollback from automation) |
| 00038 | Update client surname in Allpay                       |
