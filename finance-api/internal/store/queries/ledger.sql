-- name: CreateLedger :one
INSERT INTO ledger (id, datetime, finance_client_id, amount, notes, type, status, fee_reduction_id, created_at,
                    created_by, reference, method)
SELECT NEXTVAL('ledger_id_seq'),
       NOW(),
       fc.id,
       $2,
       $3,
       $4,
       $5,
       $6,
       NOW(),
       $7,
       gen_random_uuid(),
       ''
FROM finance_client fc
WHERE client_id = $1
RETURNING id;

-- name: CreateLedgerForCourtRef :one
INSERT INTO ledger (id, datetime, bankdate, finance_client_id, amount, notes, type, status, created_at, created_by,
                    reference, method, pis_number)
SELECT NEXTVAL('ledger_id_seq'),
       @received_date,
       @bank_date,
       fc.id,
       @amount,
       @notes,
       @type,
       @status,
       NOW(),
       @created_by,
       gen_random_uuid(),
       '',
       @pis_number
FROM finance_client fc
WHERE court_ref = @court_ref
RETURNING id;

-- name: GetLedgerForPayment :one
SELECT l.id
FROM ledger l
         JOIN finance_client fc ON fc.id = l.finance_client_id
WHERE l.amount = $1
  AND l.status = 'CONFIRMED'
  AND l.bankdate = $2
  AND l.type = $3
  AND fc.court_ref = $4
LIMIT 1;

-- name: CheckDuplicateLedger :one
SELECT EXISTS (SELECT 1
               FROM ledger l
                        JOIN finance_client fc ON fc.id = l.finance_client_id
               WHERE l.amount = @amount
                 AND l.status = 'CONFIRMED'
                 AND (COALESCE(l.pis_number, 0) <> 0 OR l.bankdate = @bank_date)
                 AND l.datetime::DATE = (@received_date::TIMESTAMP)::DATE
                 AND l.type = @type
                 AND fc.court_ref = @court_ref
                 AND COALESCE(l.pis_number, 0) = COALESCE(sqlc.narg('pis_number'), 0));
