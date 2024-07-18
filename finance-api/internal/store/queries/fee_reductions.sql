-- name: GetFeeReductions :many
SELECT fr.id,
       finance_client_id,
       type,
       startdate,
       enddate,
       datereceived,
       notes,
       deleted
FROM fee_reduction fr
         INNER JOIN finance_client fc ON fc.id = fr.finance_client_id
WHERE fc.client_id = $1
ORDER BY enddate DESC, deleted;

-- name: AddFeeReduction :one
INSERT INTO fee_reduction (id,
                           finance_client_id,
                           type,
                           startdate,
                           enddate,
                           notes,
                           datereceived,
                           created_by,
                           created_at)
VALUES (NEXTVAL('fee_reduction_id_seq'::REGCLASS),
        (SELECT id FROM finance_client WHERE client_id = $1), $2, $3, $4, $5, $6, $7, now())
RETURNING *;

-- name: CountOverlappingFeeReduction :one
SELECT COUNT(*)
FROM fee_reduction fr
         INNER JOIN finance_client fc ON fc.id = fr.finance_client_id
WHERE fc.client_id = $1
  AND fr.deleted = FALSE
  AND (fr.startdate, fr.enddate) OVERLAPS ($2, $3);

-- name: CancelFeeReduction :one
UPDATE fee_reduction
SET deleted = TRUE,  cancelled_by = $2, cancelled_at = now(), cancellation_reason = $3
WHERE id = $1
RETURNING *;

-- name: GetFeeReductionForDate :one
SELECT fr.id AS fee_reduction_id, fr.type, fr.finance_client_id
FROM fee_reduction fr
         JOIN finance_client fc ON fr.finance_client_id = fc.id
WHERE $2 >= (fr.datereceived - INTERVAL '6 months')
  AND $2 BETWEEN fr.startdate AND fr.enddate
  AND fr.deleted = FALSE
  AND fc.client_id = $1;
