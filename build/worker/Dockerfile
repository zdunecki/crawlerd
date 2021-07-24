FROM golang:1.13-alpine AS build

RUN apk update
RUN apk add --no-cache git
RUN apk add --no-cache build-base

WORKDIR /go/src/api

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN go build -o /go/bin/app cmd/worker/main.go

FROM alpine
WORKDIR /app/bin
COPY --from=build /go/bin/app app


