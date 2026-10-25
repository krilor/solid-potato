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
LEFT JOIN potato.organizations o ON t.org_number = o.org_number
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
LEFT JOIN potato.organizations o ON t.org_number = o.org_number
WHERE t.reference = sqlc.arg('reference');
