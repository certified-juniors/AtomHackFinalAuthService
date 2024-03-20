FROM golang:1.22.1-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

WORKDIR /app/cmd/film_lib

RUN go build -o main

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /app/cmd/film_lib/main /main

ENTRYPOINT ["/main"]