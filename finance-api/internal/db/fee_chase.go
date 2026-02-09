package db

import (
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// FeeChase generates a report of all outstanding debt for active clients. This is used in the debt chase process to
// chase the fee payer for payment.
// This report dynamically resizes to include all outstanding invoices for each client, up to a maximum of 23 invoices.
type FeeChase struct {
	ReportQuery
}

func NewFeeChase() ReportQuery {
	return &FeeChase{
		ReportQuery: NewReportQuery(FeeChaseQuery),
	}
}

const FeeChaseQuery = `SELECT cl.caserecnumber AS "Case_no",
				cl.id AS "Client_no",
				COALESCE(cl.salutation, '') AS "Client_title",
				cl.firstname AS "Client_forename",
				cl.surname AS "Client_surname",
				CASE WHEN do_not_invoice_warning_count.count >= 1 THEN 'Yes' ELSE 'No' END  AS "Do_not_chase",
				INITCAP(fc.payment_method) AS "Payment_method",
				COALESCE(p.deputytype, '') AS "Deputy_type",
				COALESCE(c.how_deputy_appointed, '') AS "Appt_type",
				COALESCE(dii.annualbillinginvoice, '') AS "Billing_preference",
				COALESCE(p.deputynumber::VARCHAR, '') AS "Deputy_no",
				CASE WHEN a.isairmailrequired IS TRUE THEN 'Yes' ELSE 'No' END AS "Airmail",
				COALESCE(p.salutation, '') AS "Deputy_title",
				CASE WHEN p.correspondencebywelsh IS TRUE THEN 'Yes' ELSE 'No' END AS "Deputy_Welsh",
				CASE WHEN p.specialcorrespondencerequirements_largeprint IS TRUE THEN 'Yes' ELSE 'No' END AS "Deputy_Large_Print",
				COALESCE(NULLIF(p.organisationname, ''), CONCAT(p.firstname, ' ' , p.surname)) AS "Deputy_name",
				COALESCE(p.email, '') AS "Email",
				CASE WHEN a.address_lines IS NOT NULL THEN COALESCE(a.address_lines ->> 0, '') ELSE '' END AS "Address1",
				CASE WHEN a.address_lines IS NOT NULL THEN COALESCE(a.address_lines ->> 1, '') ELSE '' END AS "Address2",
				CASE WHEN a.address_lines IS NOT NULL THEN COALESCE(a.address_lines ->> 2, '') ELSE '' END AS "Address3",
				COALESCE(a.town, '') AS "City_Town",
				COALESCE(a.county, '') AS "County",
				COALESCE(a.postcode, '') AS "Postcode",
				CONCAT('£', ABS(gi.total / 100.0)::NUMERIC(10,2)::VARCHAR(255)) AS "Total_debt",
				gi.invoice
        FROM public.persons cl
                INNER JOIN supervision_finance.finance_client fc ON cl.id = fc.client_id
                LEFT JOIN public.persons p ON cl.feepayer_id = p.id
                LEFT JOIN LATERAL (
                       SELECT a.* FROM public.addresses a WHERE a.person_id = p.id ORDER BY a.id DESC LIMIT 1
                ) a ON TRUE
                LEFT JOIN LATERAL (
                    SELECT c.howdeputyappointed AS how_deputy_appointed
					FROM supervision.order_deputy od
					INNER JOIN public.cases c ON od.order_id = c.id
					WHERE od.deputy_id = p.id
					AND c.client_id = cl.id
					AND c.orderstatus = 'ACTIVE'
					AND od.statusoncaseoverride IS NULL
					ORDER BY c.casesubtype DESC LIMIT 1
                ) c ON TRUE
                LEFT JOIN LATERAL (
                	SELECT dii.annualbillinginvoice 
                    FROM supervision.deputy_important_information dii
			    	WHERE dii.deputy_id = cl.feepayer_id
                    LIMIT 1
            	) dii ON TRUE
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
            WHERE pw.person_id = cl.id
              AND w.systemstatus = TRUE
              AND w.warningtype = 'Do not invoice'
            ) AS do_not_invoice_warning_count
            WHERE cl.type = 'actor_client'
			AND cl.clientstatus = 'ACTIVE'
			AND gi.total IS NOT NULL AND gi.total > 0;`

func (f *FeeChase) GetHeaders() []string {
	maxInvoiceCount := 23

	headers := []string{
		"Case_no",
		"Client_no",
		"Client_title",
		"Client_forename",
		"Client_surname",
		"Do_not_chase",
		"Payment_method",
		"Deputy_type",
		"Appt_type",
		"Billing_preference",
		"Deputy_no",
		"Airmail",
		"Deputy_title",
		"Deputy_Welsh",
		"Deputy_Large_Print",
		"Deputy_name",
		"Email",
		"Address1",
		"Address2",
		"Address3",
		"City_Town",
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

func (f *FeeChase) GetCallback() func(row pgx.CollectableRow) ([]string, error) {
	return func(row pgx.CollectableRow) ([]string, error) {
		var stringRow []string
		values, err := row.Values()
		if err != nil {
			return nil, err
		}

		for index, value := range values {
			if index == 24 {
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
