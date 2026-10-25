-- name: ListTransactions :many
SELECT
    reference,
    card_number,
    card_type,
    transaction_time,
    org_number,
    quantity_ore,
    status
FROM potato.transactions
where coalesce(sqlc.narg('from'), '-infinity'::timestamptz) <= transaction_time
and coalesce(sqlc.narg('to'), 'infinity'::timestamptz) > transaction_time
and coalesce(sqlc.narg('status'), status) = status;

-- name: ReadTransactionByReference :one
SELECT
    reference,
    card_number,
    card_type,
    transaction_time,
    org_number,
    quantity_ore,
    status
FROM potato.transactions
WHERE reference = sqlc.arg('reference');
