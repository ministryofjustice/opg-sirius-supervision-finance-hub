-- name: GetFeeReductions :many
SELECT fr.id,
       finance_client_id,
       type,
       startdate,
       enddate,
       datereceived,
       notes,
       deleted
FROM supervision_finance.fee_reduction fr
         INNER JOIN supervision_finance.finance_client fc ON fc.id = fr.finance_client_id
WHERE fc.client_id = $1
ORDER BY enddate DESC, deleted;

-- name: AddFeeReduction :one
INSERT INTO supervision_finance.fee_reduction (id,
                                               finance_client_id,
                                               type,
                                               startdate,
                                               enddate,
                                               notes,
                                               deleted,
                                               datereceived)
VALUES (NEXTVAL('supervision_finance.fee_reduction_id_seq'::REGCLASS),
        (SELECT id FROM supervision_finance.finance_client WHERE client_id = $1), $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: CountOverlappingFeeReduction :one
SELECT COUNT(*)
FROM supervision_finance.fee_reduction fr
         INNER JOIN supervision_finance.finance_client fc ON fc.id = fr.finance_client_id
WHERE fc.client_id = $1
  AND fr.deleted = FALSE
  AND (fr.startdate, fr.enddate) OVERLAPS ($2, $3);

-- name: CancelFeeReduction :one
UPDATE supervision_finance.fee_reduction
SET deleted = TRUE
WHERE id = $1
RETURNING *;