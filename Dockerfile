FROM golang:1.13-alpine AS build

RUN apk update
RUN apk add --no-cache git
RUN apk add --no-cache build-base

WORKDIR /app
COPY . .

RUN go build -o ./bin/api main.go
RUN go build -o ./bin/scheduler cmd/scheduler/main.go
RUN go build -o ./bin/worker cmd/worker/main.go

FROM alpine
WORKDIR /app
COPY --from=build /app/bin/api api
COPY --from=build /app/bin/scheduler scheduler
COPY --from=build /app/bin/worker worker


