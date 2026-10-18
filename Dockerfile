FROM golang:1.27-alpine AS goose
ARG GOOSE_VERSION=v3.29.0
RUN CGO_ENABLED=0 go install -trimpath -ldflags="-s -w" \
    -tags='no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb' \
    github.com/pressly/goose/v3/cmd/goose@${GOOSE_VERSION}

FROM gcr.io/distroless/static-debian12:nonroot AS migrate
COPY --from=goose /go/bin/goose /goose
COPY migrations /migrations
ENTRYPOINT ["/goose"]

FROM golang:1.27-alpine AS build
WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /app ./cmd/app

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /app /app
ENTRYPOINT ["/app"]
