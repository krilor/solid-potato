//go:generate oapi-codegen -o ./api.gen.go -config ./.oapi-config.yml ./openapi.yml
//go:generate sqlc -f .sqlc.yml generate
package main
