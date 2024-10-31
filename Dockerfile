# syntax=docker/dockerfile:1

FROM golang:1.22.6 AS build-stage
WORKDIR /app
COPY . .
RUN make build

FROM build-stage AS test-stage
RUN make test

FROM test-stage AS deploy-stage
EXPOSE 8080
ENTRYPOINT ["./bin/api"]

