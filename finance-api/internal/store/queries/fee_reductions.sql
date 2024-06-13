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

-- name: GetFeeReduction :one
select id,
       finance_client_id,
       type,
       startdate,
       enddate,
       datereceived,
       notes,
       deleted
from fee_reduction fr
where id = $1;

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