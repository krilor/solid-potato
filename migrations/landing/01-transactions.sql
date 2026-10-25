-- for transaksjoner.csv
-- minimum validation - yolo bronze style
CREATE UNLOGGED TABLE IF NOT EXISTS landing.transactions_raw (
    referanse UUID NOT NULL,
    kortnummer TEXT NOT NULL,
    korttype TEXT NOT NULL,
    dato TIMESTAMPTZ NOT NULL,
    organisasjonsnummer TEXT NOT NULL,
    belop_ore BIGINT NOT NULL,
    transaksjonsstatus TEXT NOT NULL
);

-- tiny cleanup and transformation
CREATE OR REPLACE VIEW landing.transactions AS
SELECT
    referanse AS reference,
    dato AS transaction_time,
    belop_ore AS quantity_ore,
    btrim(kortnummer) AS card_number,
    CASE lower(btrim(korttype))
        WHEN 'amex' THEN 'AmericanExpress'
        WHEN 'visa' THEN 'Visa'
        WHEN 'mastercard' THEN 'Mastercard'
        ELSE btrim(korttype)
    END AS card_type,
    btrim(organisasjonsnummer) AS org_number,
    CASE upper(btrim(transaksjonsstatus))
        WHEN 'GODKJENT' THEN 'APPROVED'
        WHEN 'AVVIST' THEN 'REJECTED'
        WHEN 'VENTENDE' THEN 'PENDING'
        ELSE upper(btrim(transaksjonsstatus))
    END AS status
FROM (
    -- assuming that we only need the latest transaction per reference
    SELECT DISTINCT ON (referanse) *
    FROM landing.transactions_raw
    ORDER BY referanse ASC, dato DESC
) AS t_uniq;

CREATE OR REPLACE FUNCTION landing.transactions_empty()
RETURNS VOID
AS $$
BEGIN
    TRUNCATE TABLE landing.transactions_raw;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION landing.transactions_merge()
RETURNS VOID
SECURITY DEFINER
AS $$
BEGIN
    INSERT INTO potato.transactions (reference, card_number, card_type, transaction_time, org_number, quantity_ore, status)
    SELECT
        reference,
        card_number,
        card_type,
        transaction_time,
        org_number,
        quantity_ore,
        status
    FROM landing.transactions
    ON CONFLICT (reference) DO UPDATE
    SET card_number = EXCLUDED.card_number,
        card_type = EXCLUDED.card_type,
        transaction_time = EXCLUDED.transaction_time,
        org_number = EXCLUDED.org_number,
        quantity_ore = EXCLUDED.quantity_ore,
        status = EXCLUDED.status;
END;
$$ LANGUAGE plpgsql;
