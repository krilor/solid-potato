# Development

## Requires

* just - command runner
* mise-en-place - for local tools
* docker-compose - for running database
* curl - for getting brreg data
* psql - for connecting to the pg database

## Running the API

The API is configured with environment variables:

| Variable      | Default     | Description                    |
| ------------- | ----------- | ------------------------------ |
| `LISTEN_ADDR` | `:8080`     | Address the HTTP server binds  |
| `DB_HOST`     | `localhost` | Database host                  |
| `DB_PORT`     | `5432`      | Database port                  |
| `DB_NAME`     | `potato`    | Database name                  |
| `DB_USER`     | `reader`    | Database user                  |
| `DB_PASSWORD` | (required)  | Database password              |
| `DB_SSLMODE`  | `prefer`    | Postgres `sslmode`             |

Locally, run it with `just run`.
