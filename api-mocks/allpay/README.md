Manual testing with the Allpay Mock
===
Imposter has been configured to provide stock responses for all queries, with optional responses based on certain
request parameters. This can be useful for testing failure conditions or end-to-end processes where the tester is not
directly in control of the data, such as event-driven API calls.

This readme contains the currently configured parameters that can be used in testing, either locally or in environments
where
the mock is deployed:

Surname parameters
---
To trigger these responses, include the value in the client's surname, e.g. update the client in Sirius to change the
surname from "Smith" to "Smith_allpay_mandate_validation".

| Value                        | Function        | Trigger                 |
|------------------------------|-----------------|-------------------------|
| `allpay_mandate_validation`  | Create Mandate  | Validation response     |
| `allpay_mandate_fail`        | Create Mandate  | Allpay failure response |
| `allpay_schedule_validation` | Create Schedule | Validation response     |
| `allpay_schedule_fail`       | Create Schedule | Allpay failure response |
| `allpay_cancel_fail`         | Cancel Mandate  | Allpay failure response |

Modulus check
---
To trigger modulus check responses in the Setup Direct Debit form, use the following inputs:

| Field          | Value    | Trigger                |
|----------------|----------|------------------------|
| Account number | 99999999 | Invalid combination    |
| Sort code      | 99-99-99 | Invalid sort code      |
| Sort code      | 00-00-00 | Non-existent sort code |

Failed payments
---
To test failed payments, three things need to happen:

1. A mandate is set up for a client with a payment schedule
2. The scheduled payment has been collected
3. The failed payment response from Allpay contains the payment details

To enable this, data is captured when a schedule is created, and can be returned as a failed payment by sending the
following event via EventBridge:

```json
{
    "source": "opg.supervision.infra",
    "detail-type": "scheduled-event",
    "detail": {
        "trigger": "failed-direct-debit-collections",
        "override": {
            "date": "2000-01-01"
        }
    }
}
```
