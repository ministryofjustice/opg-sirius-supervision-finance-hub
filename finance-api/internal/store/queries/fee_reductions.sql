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
insert into fee_reduction (finance_client_id,
                           type,
                           evidencetype,
                           startdate,
                           enddate,
                           notes,
                           deleted,
                           datereceived) values ($1, $2, $3, $4, $5, $6, $7, $8) returning *;