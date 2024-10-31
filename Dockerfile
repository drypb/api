# syntax=docker/dockerfile:1

FROM golang:1.22.6 AS build-stage
WORKDIR /app
COPY . .
RUN go build -o bin/api ./cmd/api

FROM build-stage AS test-stage
RUN go test ./...

FROM test-stage AS deploy-stage
EXPOSE 8080
ENTRYPOINT ["./bin/api"]

