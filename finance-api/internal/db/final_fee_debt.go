package db

import (
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
)

type FinalFeeDebt struct {
	ReportQuery
}

func NewFinalFeeDebt() ReportQuery {
	return &FinalFeeDebt{
		ReportQuery: NewReportQuery(FinalFeeDebtQuery),
	}
}

const FinalFeeDebtQuery = `SELECT client.caserecnumber AS "Case_no",
				client.id AS "Client_no",
				COALESCE(client.salutation, '') AS "Client_title",
				client.firstname AS "Client_forename",
				client.surname AS "Client_surname",
				TO_CHAR(mostrecentlyclosedorder.closedondate, 'YYYY-MM-DD') AS "Closed_date",
                mostrecentlyclosedorder.orderclosurereason "Closure_reason",
				CASE WHEN do_not_invoice_warning_count.count >= 1 THEN 'Yes' ELSE 'No' END  AS "Do_not_chase",
				COALESCE(fee_payer.deputytype, '') AS "Deputy_type",
				COALESCE(fee_payer.deputynumber::VARCHAR, '') AS "Deputy_no",
				CASE WHEN a.isairmailrequired IS TRUE THEN 'Yes' ELSE 'No' END AS "Airmail",
				COALESCE(fee_payer.salutation, '') AS "Deputy_title",
				COALESCE(NULLIF(fee_payer.organisationname, ''), CONCAT(fee_payer.firstname, ' ' , fee_payer.surname)) AS "Deputy_name",
				COALESCE(fee_payer.email, '') AS "Email",
				CASE WHEN a.address_lines IS NOT NULL THEN COALESCE(a.address_lines ->> 0, '') ELSE '' END AS "Address1",
				CASE WHEN a.address_lines IS NOT NULL THEN COALESCE(a.address_lines ->> 1, '') ELSE '' END AS "Address2",
				CASE WHEN a.address_lines IS NOT NULL THEN COALESCE(a.address_lines ->> 2, '') ELSE '' END AS "Address3",
				COALESCE(a.town, '') AS "City_town",
				COALESCE(a.county, '') AS "County",
				COALESCE(a.postcode, '') AS "Postcode",
				CONCAT('£', ABS(gi.total / 100.0)::NUMERIC(10,2)::VARCHAR(255)) AS "Total_debt",
				gi.invoice
        FROM public.persons client
                INNER JOIN supervision_finance.finance_client fc ON client.id = fc.client_id
                LEFT JOIN public.persons fee_payer ON client.feepayer_id = fee_payer.id
                LEFT JOIN LATERAL (
                       SELECT a.* FROM public.addresses a WHERE a.person_id = fee_payer.id ORDER BY a.id DESC LIMIT 1
                ) a ON TRUE
           , LATERAL (
            SELECT 
                JSON_AGG(
                	JSON_BUILD_OBJECT(
                   		'reference', i.reference, 
                   		'debt', CONCAT('£', ABS((i.amount - COALESCE(transactions.received, 0)) / 100.0)::NUMERIC(10,2)::VARCHAR(255))
                	) ORDER BY i.startdate
				)   AS invoice,
			   COUNT(i.reference)                                                                         		  AS count,
			   SUM(i.amount - COALESCE(transactions.received, 0))                                                 AS total
			FROM supervision_finance.invoice i
			LEFT JOIN LATERAL (
				SELECT SUM(la.amount) AS received
				FROM supervision_finance.ledger_allocation la
					 JOIN supervision_finance.ledger l ON la.ledger_id = l.id AND l.status = 'CONFIRMED'
				WHERE la.status NOT IN ('PENDING', 'UNALLOCATED')
				AND la.invoice_id = i.id
				) transactions ON TRUE
			WHERE i.amount > COALESCE(transactions.received, 0)
			AND i.finance_client_id = fc.id
			GROUP BY i.person_id
            ) AS gi
            , LATERAL (
            SELECT COUNT(w.id)
            FROM public.person_warning pw
            INNER JOIN public.warnings w ON pw.warning_id = w.id
            WHERE pw.person_id = client.id
              AND w.systemstatus = TRUE
              AND w.warningtype = 'Do not invoice'
            ) AS do_not_invoice_warning_count
        LEFT JOIN LATERAL (
                SELECT c.closedondate, c.orderclosurereason
                FROM public.cases c
                WHERE client.id = c.client_id
                AND c.closedondate IS NOT NULL
                ORDER BY c.closedondate DESC
                LIMIT 1
            ) AS mostrecentlyclosedorder ON TRUE
            WHERE client.type = 'actor_client'
			AND client.clientstatus NOT IN ('OPEN', 'ACTIVE')
			AND gi.total IS NOT NULL AND gi.total > 0;`

func (f *FinalFeeDebt) GetHeaders() []string {
	maxInvoiceCount := 23

	headers := []string{
		"Case_no",
		"Client_no",
		"Client_title",
		"Client_forename",
		"Client_surname",
		"Closed_date",
		"Closure_reason",
		"Do_not_chase",
		"Deputy_type",
		"Deputy_no",
		"Airmail",
		"Deputy_title",
		"Deputy_name",
		"Email",
		"Address1",
		"Address2",
		"Address3",
		"City_town",
		"County",
		"Postcode",
		"Total_debt",
	}

	for i := 1; i <= maxInvoiceCount; i++ {
		headers = append(headers, fmt.Sprintf("Invoice%d", i))
		headers = append(headers, fmt.Sprintf("Amount%d", i))
	}

	return headers
}

func (f *FinalFeeDebt) GetCallback() func(row pgx.CollectableRow) ([]string, error) {
	return func(row pgx.CollectableRow) ([]string, error) {
		var stringRow []string
		values, err := row.Values()
		if err != nil {
			return nil, err
		}

		for index, value := range values {
			if index == 21 {
				jsonBytes, err := json.Marshal(value)
				if err != nil {
					return nil, fmt.Errorf("error marshaling to JSON: %v", err)
				}

				var invoices []struct {
					Reference string `json:"reference"`
					Debt      string `json:"debt"`
				}

				if err := json.Unmarshal(jsonBytes, &invoices); err != nil {
					return nil, fmt.Errorf("error unmarshaling to struct: %v", err)
				}

				for _, invoice := range invoices {
					stringRow = append(stringRow, fmt.Sprintf("%v", invoice.Reference))
					stringRow = append(stringRow, fmt.Sprintf("%v", invoice.Debt))
				}
			} else {
				stringRow = append(stringRow, fmt.Sprintf("%v", value))
			}

		}
		return stringRow, nil
	}
}
