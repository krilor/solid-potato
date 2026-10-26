# bake the potato
FROM golang:1.27.3-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /potato .

FROM gcr.io/distroless/static:latest AS runtime

WORKDIR /

COPY --from=build /potato /potato

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/potato"]
