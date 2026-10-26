# Solid Potato

[Ta en potet!](https://www.nrk.no/video/portveien-2---ta-en-potet_257213)

This project is MVP for a RESTful read only API based on data from two sources.

* PostgreSQL w/ docker compose
* ELT done via curl and psql (with csv `\copy` ) to unlogged tables + simple
  merge into a target
* Backend in go with pgx
* Spec first OpenAPI document serves as the input to `oapi-codegen` which
  outputs DTOs/models
* `sqlc` for type safe code generation from sql queries
* Go stdlib in the middle

## Try it out

To run the project you need docker compose.
You also need the `just` command runner and `psql` cli tool.

```bash
brew install libpq
brew install just
```

Clone the project.
Then you can fire it up, migrate and load data.

```bash
# build the backend
docker compose build

# start all the things
docker compose up -d

# run migrations
just migrate

# load data
just load
```

The api should now be serving you data at `localhost:8080`.

```bash
curl localhost:8080/transactions
```

## TODOs and discussion points

Stuff I should have be done but did not take the time to do or adjust. Or simply
left out since this is a toy project.

* Decide on `v0` or `v1` in the API using path based versioning. Prefer `/v1/`.
* List pagination with `limit` and `offset`.
* Consider using nested response structure with `meta` and `data` elements.
* Wondering if `/reports/transactions_per_organization` is a better resource for
  the stats endpoint
* Authentication. Prefer Bearer JWT from an external STS, with scopes.
* Return `csv` and/or `gzip`ed data based on content negotiation headers.
* Test coverage on the API with testcontainers in CI
* More check constraints in the database. Both referential integrity and
  data validation.
* Conversion from øre to nok in ELT. Its a bit late to do it in the sqlc queries
  since the whole api is in NOK.
* Discuss monetary-value as float with stakeholders. Prefer int or possibly
  numeric as a string.

## License

All code and scripts as well as code examples and snippets in documentation are
licensed under [Zero-Clause BSD](./LICENSE-0BSD). This means you can copy and
use it however you want 🥳 - no strings attached. Documentation and the rest of
the repository is licensed under
[Creative Commons Attribution 4.0 International](./LICENSE-CC-BY-4.0). This
means you can copy, redistribute and adapt it, but it requires attribution.
