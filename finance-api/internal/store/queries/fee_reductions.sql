-- name: GetFeeReductions :many
select fr.id,
       finance_client_id,
       type,
       startdate,
       enddate,
       datereceived,
       notes,
       deleted
from fee_reduction fr
         inner join finance_client fc on fc.id = fr.finance_client_id
where fc.client_id = $1
order by enddate desc, deleted;

-- name: AddFeeReduction :one
insert into fee_reduction (id,
                           finance_client_id,
                           type,
                           startdate,
                           enddate,
                           notes,
                           deleted,
                           datereceived) values (nextval('fee_reduction_id_seq'::regclass), (select id from finance_client where client_id = $1), $2, $3, $4, $5, $6, $7) returning *;

-- name: CountOverlappingFeeReduction :one
SELECT COUNT(*)
from fee_reduction fr
         inner join finance_client fc on fc.id = fr.finance_client_id
where fc.client_id = $1 and fr.deleted = false
  and (fr.startdate, fr.enddate) OVERLAPS ($2, $3);

-- name: CancelFeeReduction :one
update fee_reduction set deleted = true where id = $1 returning *;

-- name: GetFeeReductionByInvoiceId :one
SELECT fr.id AS fee_reduction_id, fr.type, fr.finance_client_id
FROM invoice i
         JOIN fee_reduction fr
              ON i.finance_client_id = fr.finance_client_id
WHERE i.raiseddate >= (fr.datereceived - interval '6 months')
  AND i.raiseddate BETWEEN fr.startdate AND fr.enddate
  AND fr.id in (SELECT fere.id
                FROM fee_reduction fere
                         JOIN finance_client fc on fere.finance_client_id = fc.id
                WHERE fc.client_id = $1)
  AND fr.deleted = false
  AND i.id = $2;