CREATE TABLE IF NOT EXISTS potato.transactions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    reference UUID NOT NULL,
    card_number TEXT NOT NULL,
    card_type TEXT NOT NULL,
    transaction_time TIMESTAMPTZ NOT NULL,
    org_number TEXT NOT NULL,
    quantity_ore BIGINT NOT NULL,
    transaction_status TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_reference
ON potato.transactions (reference);

-- TODO more dq checks - not gonna spend time it

CREATE TRIGGER suppress_redundant_updates
BEFORE UPDATE ON potato.transactions
FOR EACH ROW
EXECUTE FUNCTION suppress_redundant_updates_trigger();
