# Initialize the project
init:
    pre-commit install

# Lint all-the-things
lint:
    pre-commit run --all-files
    vacuum lint openapi.yml

# Start the local development environment
start:
    docker compose up -d

# remove containers
down:
    docker compose down

# Run the API with hot reloading
serve:
    docker compose stop backend
    air

# Run unit-tests
utest:
    go test ./...

# Run integraton-tests against running server
itest host="http://localhost:8080":
    hurl --test --variable host={{host}} hurl/*.hurl


# Run fuzz-tests
ftest:
    uvx schemathesis run http://localhost:8080/openapi.json

# Drop db schemas to start over
reset:
    #!/usr/bin/env bash
    set -euo pipefail

    PGPASSWORD=postgres_pwd psql -h localhost -p 5432 -U postgres -d potato -f local/db/reset.sql

# Generate code with deterministic tools
generate:
    #!/usr/bin/env bash
    set -euo pipefail

    go generate ./...

# Run db migrations
migrate:
    #!/usr/bin/env bash
    set -euo pipefail

    for f in migrations/potato/*.sql; do
        echo "Applying migration $f"
        PGPASSWORD=potato_pwd psql -h localhost -p 5432 -U potato -d potato -f "$f"
    done

    for f in migrations/landing/*.sql; do
        echo "Applying migration $f"
        PGPASSWORD=potato_pwd psql -h localhost -p 5432 -U potato -d potato -f "$f"
    done

# Load transaction and organization data
load: _load-transaksjoner _load-enheter

_load-transaksjoner:
    #!/usr/bin/env bash
    set -euo pipefail

    PGPASSWORD=loader_pwd psql -h localhost -p 5432 -U loader -d potato -c "SELECT landing.transactions_empty()"

    PGPASSWORD=loader_pwd psql -h localhost -p 5432 -U loader -d potato -c "\copy landing.transactions_raw FROM './case/transaksjoner.csv' WITH (FORMAT csv, HEADER true)"

    PGPASSWORD=loader_pwd psql -h localhost -p 5432 -U loader -d potato -c "SELECT landing.transactions_merge()"


_load-enheter:
    #!/usr/bin/env bash
    set -euo pipefail

    # Doing a simple local cache of the Brreg data for dev speedup
    if [ ! -f case/.enheter.csv ]; then
        curl --compressed \
            'https://data.ppe.brreg.no/enhetsregisteret/api/enheter/lastned/csv' \
            -H 'accept: application/vnd.brreg.enhetsregisteret.enhet.v2+gzip' > case/.enheter.full.csv.gz

        gzip -d case/.enheter.full.csv.gz

        mlr --csv cut -f organisasjonsnummer,navn case/.enheter.full.csv > case/.enheter.csv

    else
        echo "Using cached Brreg data"
        echo "You can remove it with:"
        echo "rm case/.enheter.*"
    fi

    PGPASSWORD=loader_pwd psql -h localhost -p 5432 -U loader -d potato -c "SELECT landing.organizations_empty()"

    PGPASSWORD=loader_pwd psql -h localhost -p 5432 -U loader -d potato -c "\copy landing.organizations_raw FROM './case/.enheter.csv' WITH (FORMAT csv, HEADER true)"

    PGPASSWORD=loader_pwd psql -h localhost -p 5432 -U loader -d potato -c "SELECT landing.organizations_merge()"
