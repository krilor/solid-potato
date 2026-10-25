-- landing grants
GRANT USAGE ON SCHEMA landing TO loader;

GRANT ALL ON landing.transactions_raw TO loader;
GRANT EXECUTE ON FUNCTION landing.transactions_empty() TO loader;
GRANT EXECUTE ON FUNCTION landing.transactions_merge() TO loader;

GRANT ALL ON landing.organizations_raw TO loader;
GRANT EXECUTE ON FUNCTION landing.organizations_empty() TO loader;
GRANT EXECUTE ON FUNCTION landing.organizations_merge() TO loader;
