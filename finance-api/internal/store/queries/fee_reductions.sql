-- name: GetFeeReductions :many
select fr.id,
       discounttype,
       startdate,
       enddate,
       datereceived,
       case
           when deleted = true then 'Cancelled'
           when (startdate < CURRENT_DATE and enddate > CURRENT_DATE) then 'Active'
           when CURRENT_DATE > enddate then 'Expired'
           else '' end as status,
       notes
from fee_reduction fr
         inner join finance_client fc on fc.id = fr.finance_client_id
where fr.finance_client_id = $1
order by enddate desc, deleted;