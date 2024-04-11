-- name: GetFeeReductions :many
select fr.id,
       discounttype,
       startdate,
       enddate,
       datereceived,
       notes,
       deleted
from fee_reduction fr
         inner join finance_client fc on fc.id = fr.finance_client_id
where fr.finance_client_id = $1
order by enddate desc, deleted;