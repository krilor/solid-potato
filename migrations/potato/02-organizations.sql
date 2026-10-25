CREATE TABLE IF NOT EXISTS potato.organizations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    org_number TEXT NOT NULL,
    full_name TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_organizations_org_number
ON potato.organizations (org_number);

-- TODO more dq checks - not gonna spend time it

CREATE TRIGGER suppress_redundant_updates
BEFORE UPDATE ON potato.organizations
FOR EACH ROW
EXECUTE FUNCTION suppress_redundant_updates_trigger();
