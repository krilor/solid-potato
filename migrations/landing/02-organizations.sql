-- for enheter.csv
-- minimum validation - yolo bronze style
CREATE UNLOGGED TABLE IF NOT EXISTS landing.organizations_raw (
    organisasjonsnummer TEXT NOT NULL,
    navn TEXT NOT NULL
);

CREATE OR REPLACE VIEW landing.organizations AS
SELECT
    btrim(organisasjonsnummer) AS org_number,
    btrim(navn) AS full_name
FROM landing.organizations_raw;

CREATE OR REPLACE FUNCTION landing.organizations_empty()
RETURNS VOID
AS $$
BEGIN
    TRUNCATE TABLE landing.organizations_raw;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION landing.organizations_merge()
RETURNS VOID
SECURITY DEFINER
AS $$
BEGIN
    INSERT INTO potato.organizations (org_number, full_name)
    SELECT
        org_number,
        full_name
    FROM landing.organizations
    ON CONFLICT (org_number) DO UPDATE
    SET full_name = EXCLUDED.full_name;
END;
$$ LANGUAGE plpgsql;
