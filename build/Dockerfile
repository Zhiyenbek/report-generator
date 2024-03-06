# syntax = docker/dockerfile:1.0-experimental
FROM golang:1.17 AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -mod vendor -ldflags '-w -s' -o /app/prog -v ./cmd/main.go

#FROM migrate/migrate

#COPY ./schema /migrations

WORKDIR /app
COPY --from=build /app/prog .

ENTRYPOINT ["/app/prog"]
