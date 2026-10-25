-- api grants
GRANT USAGE ON SCHEMA potato TO reader;

GRANT SELECT ON potato.transactions TO reader;
GRANT SELECT ON potato.organizations TO reader;
