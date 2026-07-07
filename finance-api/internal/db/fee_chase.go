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
				invoice_26_af_letter.raiseddate,
				invoice_26_af_letter.templateid, 
				invoice_26_af_letter.createddate,
				any_a2_or_a6_on_case.systemtype,
				any_a2_or_a6_on_case.publisheddate,
				any_a2_or_a6_on_case.case_id,
				CONCAT('£', ABS(gi.total / 100.0)::NUMERIC(10,2)::VARCHAR(255)) AS "Total_debt",
				gi.invoice
        FROM public.persons cl
                INNER JOIN supervision_finance.finance_client fc ON cl.id = fc.client_id
                LEFT JOIN public.persons p ON cl.feepayer_id = p.id
                LEFT JOIN LATERAL (
                       SELECT a.* FROM public.addresses a WHERE a.person_id = p.id ORDER BY a.id DESC LIMIT 1
                ) a ON TRUE
                LEFT JOIN LATERAL (
                    SELECT c.howdeputyappointed AS how_deputy_appointed, c.id AS case_id
					FROM supervision.order_deputy od
					INNER JOIN public.cases c ON od.order_id = c.id
					WHERE od.deputy_id = p.id
					AND c.client_id = cl.id
					AND c.orderstatus = 'ACTIVE'
					AND od.statusoncaseoverride IS NULL
					ORDER BY c.casesubtype DESC LIMIT 1
                ) c ON TRUE
                LEFT JOIN LATERAL (
                    SELECT i2.raiseddate, ies.templateid, ies.createddate
                    FROM supervision_finance.invoice i2
                    INNER JOIN supervision_finance.invoice_email_status ies ON ies.invoice_id = i2.id
                    WHERE i2.finance_client_id = fc.id
                    AND i2.raiseddate > '2026-03-30'
					AND i2.raiseddate < '2026-04-01'
                    AND ies.templateid IN ('af1', 'AF1', 'af2', 'AF2', 'af3', 'AF3')
                ) invoice_26_af_letter ON TRUE
                LEFT JOIN LATERAL (
                    SELECT d.systemtype, d.publisheddate, c.case_id
                    FROM public.caseitem_document cd
                    INNER JOIN public.documents d ON cd.document_id = d.id
                    WHERE cd.caseitem_id = c.case_id
                    AND d.systemtype IN ('a2', 'a6')
                    AND deletedat IS NULL
                    AND publisheddate IS NOT NULL
					AND publisheddate >= '2026-04-01'
                    AND publisheddate < '2026-05-26'
                    ORDER BY publisheddate DESC LIMIT 1
                ) any_a2_or_a6_on_case ON TRUE
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
				FROM supervision.warnings w
				WHERE w.client_id = cl.id
				  AND w.isactive = TRUE
				  AND w.warningtype = 'Do not invoice'
			) AS do_not_invoice_warning_count
            WHERE cl.type = 'actor_client'
			AND cl.clientstatus = 'ACTIVE'
            AND cl.caserecnumber NOT IN (
            '20026298',
			'14253064',
			'20026464',
			'20019201',
			'20022316',
			'20017743',
			'20013743',
			'20013522',
			'20007583',
			'20023109',
			'20004224',
			'14078222',
			'20014479',
			'20007021',
			'20011268',
			'20007802',
			'20018025',
			'14266961',
			'13492565',
			'20007696',
			'20021711',
			'20026800',
			'20013874',
			'20026851',
			'20026864',
			'20012255',
			'20026717',
			'20021785',
			'20019406',
			'13880310',
			'20026852',
			'20024175',
			'13772634',
			'20025781',
			'20027001',
			'20026814',
			'20029846',
			'20015699',
			'20026992',
			'20011510',
			'20026981',
			'20026986',
			'13907804',
			'20025897',
			'20027005',
			'20020434',
			'20025911',
			'13115500',
			'20025668',
			'20025194',
			'20026374',
			'20027043',
			'20027020',
			'20006327',
			'20026029',
			'13561927',
			'20004717',
			'12153794',
			'20030046',
			'20019307',
			'20016346',
			'13106566',
			'20021872',
			'20013489',
			'20015036',
			'20030052',
			'20013770',
			'20001249',
			'20025980',
			'20023427',
			'20010830',
			'20024386',
			'20017891',
			'20011522',
			'14235363',
			'20029998',
			'20014554',
			'20027235',
			'20027249',
			'20026258',
			'20006455',
			'20011155',
			'20030719',
			'20030295',
			'20011315',
			'14089289',
			'20030304',
			'14241247',
			'14152022',
			'20010495',
			'20013243',
			'20010029',
			'20011064',
			'20021605',
			'20018468',
			'14189852',
			'20027239',
			'20030292',
			'20030240',
			'20025954',
			'20027209',
			'20026712',
			'14207742',
			'20013700',
			'20023121',
			'20018238',
			'20026350',
			'20030256',
			'1421922T',
			'14264880',
			'20025735',
			'20022281',
			'20018662',
			'20022313',
			'20025766',
			'20022872',
			'20030103',
			'20018633',
			'20026837',
			'20027003',
			'20022635',
			'20010504',
			'20016059',
			'20028523',
			'20015914',
			'20013230',
			'20018485',
			'20013721',
			'20020821',
			'20014940',
			'20030262',
			'20018684',
			'20024812',
			'20030338',
			'20023137',
			'20022990',
			'13444465',
			'20022310',
			'20017971',
			'20023523',
			'20020656',
			'20018615',
			'20026320',
			'20010807',
			'20030170',
			'20008076',
			'20030155',
			'20017020',
			'20025133',
			'20022087',
			'20027486',
			'20024153',
			'20027169',
			'20025124',
			'20006980',
			'14070738',
			'13288990',
			'20027205',
			'20022676',
			'20023598',
			'20024875',
			'13427871',
			'13034753',
			'20029972',
			'1347806T',
			'20018545',
			'20015868',
			'20025979',
			'20024735',
			'20021868',
			'20020312',
			'20027305',
			'20016974',
			'20025412')
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
		"invoice_26_af_letter.raiseddate",
		"invoice_26_af_letter.systemtype",
		"invoice_26_af_letter.publisheddate",
		"any_a2_or_a6_on_case.systemtype",
		"any_a2_or_a6_on_case.createddate",
		"any_a2_or_a6_on_case.case_id",
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
			if index == 30 {
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

