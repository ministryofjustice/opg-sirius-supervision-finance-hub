package db

type FeeChase struct{}

const FeeChaseQuery = `SELECT cl.caserecnumber AS "Case_no",
               cl.id AS "Client_no",
               COALESCE(cl.salutation, '') AS "Client_title",
               cl.firstname AS "Client_forename",
               cl.surname AS "Client_surname",
               CASE WHEN do_not_invoice_warning_count.count >= 1 THEN 'Yes' ELSE 'No' END  AS "Do_not_chase",
               initcap(fc.payment_method) AS "Payment_method",
               COALESCE(p.deputytype, '') AS "Deputy_type",
               COALESCE(p.deputynumber::VARCHAR, '') AS "Deputy_no",
               CASE WHEN a.isairmailrequired IS TRUE THEN 'Yes' ELSE 'No' END AS "Airmail",
               COALESCE(p.salutation, '') AS "Deputy_title",
               CASE WHEN p.correspondencebywelsh IS TRUE THEN 'Yes' ELSE 'No' END AS "Deputy_Welsh",
               CASE WHEN p.specialcorrespondencerequirements_largeprint IS TRUE THEN 'Yes' ELSE 'No' END AS "Deputy_Large_Print",
               COALESCE(nullif(p.organisationname, ''), CONCAT(p.firstname, ' ' , p.surname)) AS "Deputy_name",
               COALESCE(p.email, '') AS "Email",
               CASE WHEN a.address_lines IS NOT NULL THEN COALESCE(a.address_lines ->> 0, '') ELSE '' END AS "Address1",
               CASE WHEN a.address_lines IS NOT NULL THEN COALESCE(a.address_lines ->> 1, '') ELSE '' END AS "Address2",
               CASE WHEN a.address_lines IS NOT NULL THEN COALESCE(a.address_lines ->> 2, '') ELSE '' END AS "Address3",
               COALESCE(a.town, '') AS "City_Town",
               COALESCE(a.county, '') AS "County",
               COALESCE(a.postcode, '') AS "Postcode",
               ABS(gi.total / 100.0)::NUMERIC(10,2)::VARCHAR(255) AS "Total_debt",
               (SELECT * FROM json_array_elements(gi.invoice)) AS invoice
        FROM public.persons cl
                INNER JOIN supervision_finance.finance_client fc on cl.id = fc.client_id
                LEFT JOIN public.persons p ON cl.feepayer_id = p.id
                LEFT JOIN LATERAL (
                       SELECT a.* FROM public.addresses a WHERE a.person_id = p.id ORDER BY a.id DESC LIMIT 1
                ) a ON TRUE
           , LATERAL (
            SELECT 
                json_agg(json_build_object(
                   'reference', i.reference, 'debt', ABS((i.amount - COALESCE(transactions.received, 0)) / 100.0)::NUMERIC(10,2)::VARCHAR(255)) ORDER BY i.startdate)   as invoice,
                   count(i.reference)                                                                         		  as count,
                   SUM(i.amount - COALESCE(transactions.received, 0))                                                 as total
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
            group by i.person_id, i.amount
            ) AS gi
           , LATERAL (
            SELECT COUNT(w.id)
            FROM public.person_warning pw
            INNER JOIN public.warnings w on pw.warning_id = w.id
            WHERE pw.person_id = cl.id
              AND w.systemstatus = true
              AND w.warningtype = 'Do not invoice'
            ) AS do_not_invoice_warning_count
            WHERE cl.type = 'actor_client'
              AND cl.clientstatus = 'ACTIVE'
              AND gi.total IS NOT NULL AND gi.total > 0;`

func (c *FeeChase) GetHeaders() []string {
	return []string{
		"Case_no",
		"Client_no",
		"Client_title",
		"Client_forename",
		"Client_surname",
		"Do_not_chase",
		"Payment_method",
		"Deputy_type",
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
		"Invoice",
	}
}

func (c *FeeChase) GetQuery() string {
	return FeeChaseQuery
}

func (c *FeeChase) GetParams() []any {
	return []any{}
}
