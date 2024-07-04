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

-- name: GetFeeReductionForDate :one
SELECT fr.id AS fee_reduction_id, fr.type, fr.finance_client_id
FROM fee_reduction fr
         JOIN finance_client fc on fr.finance_client_id = fc.id
WHERE $2 >= (fr.datereceived - interval '6 months')
  AND $2 BETWEEN fr.startdate AND fr.enddate
  AND fr.deleted = false
  AND fc.client_id = $1;
