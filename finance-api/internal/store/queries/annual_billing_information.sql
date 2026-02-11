SET SEARCH_PATH TO supervision_finance;
-- name: GetAnnualBillingYear :one
SELECT value FROM supervision_finance.property WHERE key = 'AnnualBillingYear' LIMIT 1;

-- name: GetAnnualBillingLettersInformation :many
SELECT
    COUNT(DISTINCT i.id) AS count,
    COALESCE(ies.status, 'UNPROCESSED') AS status
FROM supervision_finance.invoice i
    JOIN public.persons c ON i.person_id = c.id
    JOIN public.cases o ON o.client_id = i.person_id
    LEFT JOIN supervision_finance.invoice_email_status ies ON i.id = ies.invoice_id
WHERE i.startdate::DATE >= $1::DATE
  AND i.enddate::DATE <= $2::DATE
  AND i.feetype NOT IN ('N2', 'N3')
  AND (
    (
        ies.invoice_id IS NULL
      AND o.orderstatus = 'ACTIVE'
      AND c.clientstatus <> 'DEATH_NOTIFIED'
    )
    OR ies.invoice_id IS NOT NULL
)
GROUP BY ies.status;
