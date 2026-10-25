-- name: ListTransactions :many
SELECT
    t.reference,
    t.card_number,
    t.card_type,
    t.transaction_time,
    t.org_number,
    o.full_name as org_name,
    (t.quantity_ore / 100.0)::float4 as quantity,
    t.status
FROM potato.transactions t
INNER JOIN potato.organizations o ON t.org_number = o.org_number
WHERE coalesce(sqlc.narg('from'), '-infinity'::timestamptz) <= t.transaction_time
AND coalesce(sqlc.narg('to'), 'infinity'::timestamptz) > t.transaction_time
AND coalesce(sqlc.narg('status'), t.status) = t.status;

-- name: ReadTransactionByReference :one
SELECT
    t.reference,
    t.card_number,
    t.card_type,
    t.transaction_time,
    t.org_number,
    o.full_name as org_name,
    (t.quantity_ore / 100.0)::float4 as quantity,
    t.status
FROM potato.transactions t
INNER JOIN potato.organizations o ON t.org_number = o.org_number
WHERE t.reference = sqlc.arg('reference');

-- name: ListTransactionStats :many
SELECT
    t.org_number,
    o.full_name as org_name,
    (sum(t.quantity_ore) / 100.0)::float4 as sum_quantity,
    count(*) FILTER (WHERE status = 'APPROVED') as num_approved,
    count(*) FILTER (WHERE status = 'REJECTED') as num_rejected,
    count(*) FILTER (WHERE status = 'PENDING') as num_pending
FROM potato.transactions t
INNER JOIN potato.organizations o ON t.org_number = o.org_number
GROUP BY t.org_number, o.full_name;
